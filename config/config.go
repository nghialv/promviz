package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func LoadFile(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	*cfg = DefaultConfig

	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

var (
	DefaultConfig = Config{
		GraphName: "promviz",
	}

	DefaultNodeMapping = NodeMapping{
		Label:       "",
		Regex:       MustNewRegexp("(.*)"),
		Replacement: "$1",
		Class:       "default",
	}

	DefaultClass = Class{
		Name:  "default",
		Color: "rgb(186, 213, 237)",
	}
)

type Config struct {
	GraphName    string      `yaml:"graphName"`
	GlobalLevel  GlobalLevel `yaml:"globalLevel"`
	ClusterLevel []*Cluster  `yaml:"clusterLevel"`
	Classes      []*Class    `yaml:"classes,omitempty"`
}

type GlobalLevel struct {
	MaxVolumeRate float64       `yaml:"maxVolumeRate,omitempty"`
	Connections   []*Connection `yaml:"clusterConnections,omitempty"`
}

type Cluster struct {
	Cluster       string        `yaml:"cluster"`
	MaxVolumeRate float64       `yaml:"maxVolumeRate,omitempty"`
	Connections   []*Connection `yaml:"serviceConnections,omitempty"`
	NodeNotices   []*NodeNotice `yaml:"serviceNotices,omitempty"`
}

type Connection struct {
	Name          string              `yaml:"name"`
	Query         string              `yaml:"query,omitempty"`
	PrometheusURL string              `yaml:"prometheusURL,omitempty"`
	Source        *NodeMapping        `yaml:"source,omitempty"`
	Target        *NodeMapping        `yaml:"target,omitempty"`
	Status        *Status             `yaml:"status,omitempty"`
	Notices       []*ConnectionNotice `yaml:"notices,omitempty"`
}

func (c *Connection) QueryLink() string {
	promURL := strings.TrimSuffix(c.PrometheusURL, "/")
	escapedQuery := url.QueryEscape(c.Query)

	return fmt.Sprintf("%s/graph?g0.expr=%s", promURL, escapedQuery)
}

type Class struct {
	Name  string `yaml:"name"`
	Color string `yaml:"color,omitempty"`
}

type NodeNotice struct {
	Name              string            `yaml:"name"`
	Title             string            `yaml:"title"`
	SubTitle          string            `yaml:"subtitle"`
	Link              string            `yaml:"link"`
	Query             string            `yaml:"query,omitempty"`
	PrometheusURL     string            `yaml:"prometheusURL,omitempty"`
	SeverityThreshold SeverityThreshold `yaml:"severityThreshold"`
	Node              *NodeMapping      `yaml:"node,omitempty"`
}

func (nn *NodeNotice) QueryLink() string {
	if nn.Link != "" {
		return nn.Link
	}

	promURL := strings.TrimSuffix(nn.PrometheusURL, "/")
	escapedQuery := url.QueryEscape(nn.Query)
	return fmt.Sprintf("%s/graph?g0.expr=%s", promURL, escapedQuery)
}

type NodeMapping struct {
	Label       string `yaml:"label,omitempty"`
	Regex       Regexp `yaml:"regex,omitempty"`
	Replacement string `yaml:"replacement,omitempty"`
	Class       string `yaml:"class,omitempty"`
}

type Status struct {
	Label        string  `yaml:"label"`
	WarningRegex *Regexp `yaml:"warningRegex,omitempty"`
	DangerRegex  *Regexp `yaml:"dangerRegex,omitempty"`
}

type ConnectionNotice struct {
	Name              string            `yaml:"name"`
	Title             string            `yaml:"title"`
	SubTitle          string            `yaml:"subtitle"`
	Link              string            `yaml:"link"`
	StatusType        string            `yaml:"statusType"`
	SeverityThreshold SeverityThreshold `yaml:"severityThreshold"`
}

type SeverityThreshold struct {
	Info    float64 `yaml:"info,omitempty"`
	Warning float64 `yaml:"warning,omitempty"`
	Error   float64 `yaml:"error,omitempty"`
}

type Regexp struct {
	*regexp.Regexp
	Original string
}

func NewRegexp(s string) (Regexp, error) {
	regex, err := regexp.Compile(s)
	return Regexp{
		Regexp:   regex,
		Original: s,
	}, err
}

func MustNewRegexp(s string) Regexp {
	re, err := NewRegexp(s)
	if err != nil {
		panic(err)
	}
	return re
}

func (re *Regexp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	r, err := NewRegexp(s)
	if err != nil {
		return err
	}
	*re = r
	return nil
}

func (nm *NodeMapping) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*nm = DefaultNodeMapping
	type plain NodeMapping
	if err := unmarshal((*plain)(nm)); err != nil {
		return err
	}
	if nm.Label == "" && nm.Replacement == "" {
		return fmt.Errorf("Invalid node mapping")
	}
	return nil
}
