package retrieval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/nmnellis/vistio/api/config"
	"github.com/nmnellis/vistio/api/model"
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
	clusterMap := make(map[string]*config.Cluster, len(g.cfg.ClusterLevel))

	group.Go(func() error {
		cs, err := g.generateNodeConnectionSet(groupCtx, g.cfg.GlobalLevel.Connections, nil, ts, newClusterNode)
		if err != nil {
			return err
		}
		clusters = cs
		return nil
	})

	for _, cluster := range g.cfg.ClusterLevel {
		cluster := cluster
		clusterMap[cluster.Cluster] = cluster

		group.Go(func() error {
			ss, err := g.generateNodeConnectionSet(groupCtx, cluster.Connections, cluster.NodeNotices, ts, newServiceNode)
			if err != nil {
				return err
			}
			services[cluster.Cluster] = ss
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	classes := make([]*model.Class, 0, len(g.cfg.Classes))
	found := false
	for _, c := range g.cfg.Classes {
		if c.Name == "default" {
			found = true
		}
		classes = append(classes, &model.Class{
			Name:  c.Name,
			Color: c.Color,
		})
	}
	if !found {
		classes = append(classes, &model.Class{
			Name:  config.DefaultClass.Name,
			Color: config.DefaultClass.Color,
		})
	}

	for _, n := range clusters.Nodes {
		if cluster, ok := services[n.Name]; ok {
			n.Nodes = cluster.Nodes
			n.Connections = cluster.Connections
			n.MaxVolume = clusterMap[n.Name].MaxVolume // calculateMaxVolume(cluster.Nodes, cluster.Connections, clusterMap[n.Name].MaxVolumeRate)
		}
	}

	graph := &model.VizceralGraph{
		Renderer:         "global",
		Name:             g.cfg.GraphName,
		MaxVolume:        g.cfg.GlobalLevel.MaxVolume, // calculateMaxVolume(clusters.Nodes, clusters.Connections, g.cfg.GlobalLevel.MaxVolumeRate),
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

	groupNotices := make([](map[string][]*model.Notice), len(cfgNotices), len(cfgNotices))

	for i, cfgNoti := range cfgNotices {
		i, cfgNoti := i, cfgNoti
		group.Go(func() error {
			value, err := g.querier.Query(groupCtx, cfgNoti.PrometheusURL, cfgNoti.Query, ts)
			if err != nil {
				g.logger.Error("Failed to send promQuery",
					zap.Error(err),
					zap.String("prometheusURL", cfgNoti.PrometheusURL),
					zap.String("query", cfgNoti.Query),
				)
				// TODO: return nil or error
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
				if n.Class != "" && (nodeMap[n.Name].Class == "" || nodeMap[n.Name].Class == "default") {
					nodeMap[n.Name].Class = n.Class
				}
			}
		}
	}

	for i := range groupNotices {
		for k, noti := range groupNotices[i] {
			if node, ok := nodeMap[k]; ok {
				node.Notices = append(node.Notices, noti...)
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

			severity := -1
			switch {
			case notice.SeverityThreshold.Error > 0 && rate >= notice.SeverityThreshold.Error:
				severity = 2
			case notice.SeverityThreshold.Warning > 0 && rate >= notice.SeverityThreshold.Warning:
				severity = 1
			case notice.SeverityThreshold.Info > 0 && rate >= notice.SeverityThreshold.Info:
				severity = 0
			}
			if severity < 0 {
				continue
			}

			t, err := template.New("title").Parse(notice.Title)
			if err != nil {
				continue
			}

			title := notice.Title
			var buf bytes.Buffer
			labelMap := map[string]string{
				"value": fmt.Sprintf("%.5f", rate),
			}

			if err = t.Execute(&buf, labelMap); err != nil {
				g.logger.Error("Failed to execute rendering notice template",
					zap.Error(err),
					zap.String("title", title),
					zap.Any("labelMap", labelMap))
			}
			title = buf.String()
			link := notice.Link
			if link == "" {
				link = conn.QueryLink()
			}

			vconn.Notices = append(vconn.Notices, &model.Notice{
				Title:    title,
				Subtitle: notice.SubTitle,
				Link:     link,
				Severity: severity,
			})
		}

		connections = append(connections, vconn)
	}
	return connections
}

func (g *generator) generateNodeNotices(vector prommodel.Vector, noti *config.NodeNotice) map[string][]*model.Notice {
	notices := make(map[string][]*model.Notice)
	for _, s := range vector {
		logger := g.logger.With(
			zap.Any("noti", noti),
			zap.Any("sample", s))

		node, err := extractNodeName(s, noti.Service)
		if err != nil {
			logger.Warn("Could not determine node", zap.Error(err))
			continue
		}

		value := float64(s.Value)
		severity := -1

		switch {
		case noti.SeverityThreshold.Error > 0 && value >= noti.SeverityThreshold.Error:
			severity = 2
		case noti.SeverityThreshold.Warning > 0 && value >= noti.SeverityThreshold.Warning:
			severity = 1
		case noti.SeverityThreshold.Info > 0 && value >= noti.SeverityThreshold.Info:
			severity = 0
		}
		if severity < 0 {
			continue
		}

		t, err := template.New("title").Parse(noti.Title)
		if err != nil {
			logger.Warn("Failed to parse notice title", zap.Error(err))
			continue
		}

		title := noti.Title
		labelMap := make(map[string]string, len(s.Metric)+1)
		for k, v := range s.Metric {
			labelMap[string(k)] = string(v)
		}
		labelMap["value"] = fmt.Sprintf("%.5f", value)
		var buf bytes.Buffer

		if err = t.Execute(&buf, labelMap); err != nil {
			logger.Error("Failed to execute rendering notice template",
				zap.Error(err),
				zap.Any("labelMap", labelMap))
			continue
		}
		title = buf.String()

		if _, ok := notices[node]; !ok {
			notices[node] = make([]*model.Notice, 0)
		}

		notices[node] = append(notices[node], &model.Notice{
			Title:    title,
			Subtitle: noti.SubTitle,
			Link:     noti.QueryLink(),
			Severity: severity,
		})
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

	if maxVolumeRate <= 0 || maxVolumeRate > 1 {
		maxVolumeRate = 0.5
	}

	return float64(max) / maxVolumeRate
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
