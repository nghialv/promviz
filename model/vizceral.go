package model

type VizceralGraph struct {
	Renderer         string        `json:"renderer"`
	Name             string        `json:"name"`
	MaxVolume        float64       `json:"maxVolume,omitempty"`
	ServerUpdateTime int64         `json:"serverUpdateTime"`
	Nodes            []*Node       `json:"nodes"`
	Connections      []*Connection `json:"connections"`
	Classes          []*Class      `json:"classes"`
}

type Node struct {
	Name        string        `json:"name"`
	Renderer    string        `json:"renderer,omitempty"`
	DisplayName string        `json:"displayName,omitempty"`
	Updated     int64         `json:"updated,omitempty"`
	MaxVolume   float64       `json:"maxVolume,omitempty"`
	Class       string        `json:"class,omitempty"`
	Metadata    *Metadata     `json:"metadata,omitempty"`
	Nodes       []*Node       `json:"nodes,omitempty"`
	Connections []*Connection `json:"connections,omitempty"`
	Notices     []*Notice     `json:"notices,omitempty"`
	// Props
}

type Connection struct {
	Source   string    `json:"source"`
	Target   string    `json:"target"`
	Class    string    `json:"class,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
	Metrics  *Metrics  `json:"metrics,omitempty"`
	Notices  []*Notice `json:"notices,omitempty"`
}

type Metadata struct {
	Streaming int `json:"streaming"`
}

type Notice struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Link     string `json:"link,omitempty"`
	Severity int    `json:"severity,omitempty"`
}

type Metrics struct {
	Danger  float64 `json:"danger,omitempty"`
	Warning float64 `json:"warning,omitempty"`
	Normal  float64 `json:"normal,omitempty"`
}

type Class struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type NodeConnectionSet struct {
	Nodes       []*Node
	Connections []*Connection
}
