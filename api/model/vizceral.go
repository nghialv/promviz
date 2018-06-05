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
	Metadata    *Metadata     `json:"metadata"`
	Nodes       []*Node       `json:"nodes,omitempty"`
	Connections []*Connection `json:"connections"`
	Notices     []*Notice     `json:"notices"`
	// Props
}

type Connection struct {
	Source   string    `json:"source"`
	Target   string    `json:"target"`
	Class    string    `json:"class,omitempty"`
	Metadata *Metadata `json:"metadata"`
	Metrics  *Metrics  `json:"metrics"`
	Notices  []*Notice `json:"notices"`
}

type Metadata struct {
	Streaming int `json:"streaming"`
}

type Notice struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Link     string `json:"link,omitempty"`
	Severity int    `json:"severity"`
}

type Metrics struct {
	Danger  float64 `json:"danger"`
	Warning float64 `json:"warning"`
	Normal  float64 `json:"normal"`
}

type Class struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type NodeConnectionSet struct {
	Nodes       []*Node
	Connections []*Connection
}
