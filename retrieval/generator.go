package retrieval

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nghialv/promviz/config"
	"github.com/nghialv/promviz/model"
	prommodel "github.com/prometheus/common/model"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type generator struct {
	logger  *zap.Logger
	cfg     *config.Config
	querier querier
}

func (g *generator) generateSnapshot(ctx context.Context, ts time.Time) (*model.Snapshot, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	var clusters *model.NodeConnectionSet
	services := make(map[string]*model.NodeConnectionSet)

	group.Go(func() error {
		cs, err := g.generateNodeConnectionSet(groupCtx, g.cfg.ClusterConnections, nil, ts, newClusterNode)
		if err != nil {
			return err
		}
		clusters = cs
		return nil
	})

	for _, cluster := range g.cfg.Clusters {
		cluster := cluster
		group.Go(func() error {
			ss, err := g.generateNodeConnectionSet(groupCtx, cluster.ServiceConnections, cluster.ServiceNotices, ts, newServiceNode)
			if err != nil {
				return err
			}
			services[cluster.Name] = ss
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	classes := make([]*model.Class, 0, len(g.cfg.Classes))
	for _, c := range g.cfg.Classes {
		classes = append(classes, &model.Class{
			Name:  c.Name,
			Color: c.Color,
		})
	}

	for _, n := range clusters.Nodes {
		if cluster, ok := services[n.Name]; ok {
			n.Nodes = cluster.Nodes
			n.Connections = cluster.Connections
			n.MaxVolume = calculateMaxVolume(cluster.Nodes, cluster.Connections, g.cfg.MaxVolumeRate)
		}
	}

	graph := &model.VizceralGraph{
		Renderer:         "global",
		Name:             g.cfg.GraphName,
		MaxVolume:        calculateMaxVolume(clusters.Nodes, clusters.Connections, 1.5),
		ServerUpdateTime: ts.Unix(),
		Nodes:            clusters.Nodes,
		Connections:      clusters.Connections,
		Classes:          classes,
	}
	jsondata, err := json.Marshal(graph)
	if err != nil {
		return nil, err
	}

	snapshot := &model.Snapshot{
		Timestamp: ts,
		GraphJSON: string(jsondata),
	}

	return snapshot, nil
}

func (g *generator) generateNodeConnectionSet(ctx context.Context, cfgConns []*config.Connection, cfgNotices []*config.NodeNotice, ts time.Time, nodeFactory func(string) *model.Node) (*model.NodeConnectionSet, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	groupConns := make([]([]*model.Connection), len(cfgConns), len(cfgConns))

	for i, cfgConn := range cfgConns {
		i, cfgConn := i, cfgConn
		group.Go(func() error {
			value, err := g.querier.Query(groupCtx, cfgConn.PrometheusURL, cfgConn.Query, ts)
			if err != nil {
				g.logger.Error("Failed to send prom query",
					zap.Error(err),
					zap.String("prometheusURL", cfgConn.PrometheusURL),
					zap.String("query", cfgConn.Query),
				)
				return err
			}
			vector, ok := value.(prommodel.Vector)
			if !ok {
				g.logger.Info("Unexpected type", zap.Any("value", value))
				return nil
			}
			groupConns[i] = g.generateConnections(vector, cfgConn)
			return nil
		})
	}

	groupNotices := make([](map[string]*model.Notice), len(cfgNotices), len(cfgNotices))

	for i, cfgNoti := range cfgNotices {
		i, cfgNoti := i, cfgNoti
		group.Go(func() error {
			value, err := g.querier.Query(groupCtx, cfgNoti.PrometheusURL, cfgNoti.Query, ts)
			if err != nil {
				g.logger.Error("Failed to send prom query",
					zap.Error(err),
					zap.String("prometheusURL", cfgNoti.PrometheusURL),
					zap.String("query", cfgNoti.Query),
				)
				// TODO: rethink
				return nil
			}
			vector, ok := value.(prommodel.Vector)
			if !ok {
				g.logger.Info("Unexpected type", zap.Any("value", value))
				return nil
			}
			groupNotices[i] = g.generateNodeNotices(vector, cfgNoti)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		g.logger.Error("Failed to generate NodeConnectionSet", zap.Error(err))
		// TODO: should fast return or not
		// return nil, err
	}

	nodeMap := make(map[string]*model.Node)

	for i, cfgConn := range cfgConns {
		for _, conn := range groupConns[i] {
			ns := []struct {
				Name  string
				Class string
			}{
				{conn.Source, cfgConn.Source.Class},
				{conn.Target, cfgConn.Target.Class},
			}
			for _, n := range ns {
				if _, ok := nodeMap[n.Name]; !ok {
					nodeMap[n.Name] = nodeFactory(n.Name)
				}
				if n.Class != "" && (nodeMap[n.Name].Class == "" || nodeMap[n.Name].Class == "normal") {
					nodeMap[n.Name].Class = n.Class
				}
			}
		}
	}

	for i := range groupNotices {
		for k, noti := range groupNotices[i] {
			if node, ok := nodeMap[k]; ok {
				node.Notices = append(node.Notices, noti)
			}
		}
	}

	nodes := make([]*model.Node, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}

	connections := make([]*model.Connection, 0)
	for i := range groupConns {
		connections = append(connections, groupConns[i]...)
	}

	set := &model.NodeConnectionSet{
		Nodes:       nodes,
		Connections: connections,
	}

	return set, nil
}

func (g *generator) generateConnections(vector prommodel.Vector, conn *config.Connection) []*model.Connection {
	type metrics struct {
		Source  string
		Target  string
		All     float64
		Normal  float64
		Danger  float64
		Warning float64
	}
	metricMap := make(map[string]*metrics)

	for _, s := range vector {
		source, err := extractNodeName(s, conn.Source)
		if err != nil {
			g.logger.Warn("Could not determine source node",
				zap.Error(err),
				zap.Any("source", conn.Source),
				zap.Any("sample", s))
			continue
		}

		target, err := extractNodeName(s, conn.Target)
		if err != nil {
			g.logger.Warn("Could not determine target node",
				zap.Error(err),
				zap.Any("target", conn.Target), zap.Error(err),
				zap.Any("sample", s))
			continue
		}

		key := fmt.Sprintf("%s/%s", source, target)
		m, ok := metricMap[key]
		if !ok {
			m = &metrics{
				Source: string(source),
				Target: string(target),
			}
			metricMap[key] = m
		}

		m.All += float64(s.Value)
		matched := false
		if conn.Status != nil {
			if status, ok := s.Metric[prommodel.LabelName(conn.Status.Label)]; ok {
				if conn.Status.DangerRegex != nil {
					if conn.Status.DangerRegex.Match([]byte(status)) {
						m.Danger += float64(s.Value)
						matched = true
					}
				}
				if conn.Status.WarningRegex != nil && !matched {
					if conn.Status.WarningRegex.Match([]byte(status)) {
						m.Warning += float64(s.Value)
						matched = true
					}
				}
			} else {
				g.logger.Warn("Could not find status label",
					zap.String("label", conn.Status.Label),
					zap.Any("sample", s))
			}
		}
		if !matched {
			m.Normal += float64(s.Value)
		}
	}

	connections := make([]*model.Connection, 0, len(metricMap))
	for _, m := range metricMap {
		vconn := &model.Connection{
			Source: m.Source,
			Target: m.Target,
			Metadata: &model.Metadata{
				Streaming: 1,
			},
			Metrics: &model.Metrics{
				Normal:  m.Normal,
				Danger:  m.Danger,
				Warning: m.Warning,
			},
			Notices: []*model.Notice{},
		}
		for _, notice := range conn.Notices {
			rate := 0.0
			switch notice.StatusType {
			case "danger":
				rate = m.Danger / m.All
			case "warning":
				rate = m.Warning / m.All
			}

			if rate >= notice.Threshold {
				link := notice.Link
				if link == "" {
					link = conn.QueryLink()
				}
				vconn.Notices = append(vconn.Notices, &model.Notice{
					Title:    fmt.Sprintf("[%.2f] %s", rate, notice.Title),
					Subtitle: notice.SubTitle,
					Link:     link,
					Severity: notice.Severity,
				})
			}
		}
		connections = append(connections, vconn)
	}
	return connections
}

func (g *generator) generateNodeNotices(vector prommodel.Vector, noti *config.NodeNotice) map[string]*model.Notice {
	notices := make(map[string]*model.Notice)
	for _, s := range vector {
		node, err := extractNodeName(s, noti.Node)
		if err != nil {
			g.logger.Warn("Could not determine node",
				zap.Error(err),
				zap.Any("node", noti.Node),
				zap.Any("sample", s))
			continue
		}

		notices[node] = &model.Notice{
			Title:    noti.Title,
			Subtitle: noti.SubTitle,
			Link:     noti.QueryLink(),
			Severity: noti.Severity,
		}
	}
	return notices
}

func extractNodeName(sample *prommodel.Sample, mapping *config.NodeMapping) (string, error) {
	if mapping.Label == "" {
		return mapping.Replacement, nil
	}

	pv, ok := sample.Metric[prommodel.LabelName(mapping.Label)]
	if !ok {
		return "", fmt.Errorf("Not found label %s", mapping.Label)
	}
	value := string(pv)
	if value == "" {
		return "", fmt.Errorf("The value of label (%s) is empty", mapping.Label)
	}

	indexes := mapping.Regex.FindStringSubmatchIndex(value)
	res := mapping.Regex.ExpandString([]byte{}, mapping.Replacement, value, indexes)
	return string(res), nil
}

func calculateMaxVolume(nodes []*model.Node, connections []*model.Connection, maxVolumeRate float64) float64 {
	if len(nodes) == 0 {
		return 0
	}

	nodeMap := make(map[string]int, len(nodes))
	for _, c := range connections {
		if c.Metrics != nil {
			nodeMap[c.Source] += int(c.Metrics.Danger + c.Metrics.Warning + c.Metrics.Normal)
		}
	}

	max := 0
	for _, n := range nodeMap {
		if max < n {
			max = n
		}
	}
	return float64(max) * maxVolumeRate
}

func newServiceNode(name string) *model.Node {
	return &model.Node{
		Name:     name,
		Renderer: "focusedChild",
		Metadata: &model.Metadata{
			Streaming: 1,
		},
	}
}

func newClusterNode(name string) *model.Node {
	return &model.Node{
		Name:     name,
		Renderer: "region",
		Metadata: &model.Metadata{
			Streaming: 1,
		},
	}
}
