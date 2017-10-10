package config

import (
	"fmt"
	"io/ioutil"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

// TODO: maxVolumeRate

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
		Class:       "",
	}
)

type Config struct {
	GraphName          string        `yaml:"graphName"`
	Clusters           []*Cluster    `yaml:"clusters,omitempty"`
	ClusterConnections []*Connection `yaml:"clusterConnections,omitempty"`
	Classes            []*NodeClass  `yaml:"classes,omitempty"`
}

type Cluster struct {
	Name               string        `yaml:"name"`
	ServiceConnections []*Connection `yaml:"serviceConnections,omitempty"`
	ServiceNotices     []*NodeNotice `yaml:"serviceNotices,omitempty"`
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

type NodeClass struct {
	Name  string `yaml:"name"`
	Color string `yaml:"color,omitempty"`
}

type NodeNotice struct {
	Name          string       `yaml:"name"`
	Title         string       `yaml:"title"`
	SubTitle      string       `yaml:"subtitle"`
	Link          string       `yaml:"link"`
	Severity      int          `yaml:"severity"`
	Query         string       `yaml:"query,omitempty"`
	PrometheusURL string       `yaml:"prometheusURL,omitempty"`
	Node          *NodeMapping `yaml:"node,omitempty"`
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
	Name       string  `yaml:"name"`
	Title      string  `yaml:"title"`
	SubTitle   string  `yaml:"subtitle"`
	Link       string  `yaml:"link"`
	Severity   int     `yaml:"severity"`
	StatusType string  `yaml:"statusType"`
	Threshold  float64 `yaml:"threshold"`
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
