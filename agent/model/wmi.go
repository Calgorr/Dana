package model

type Wmi struct {
	ID       int         `json:"id"`
	Host     string      `json:"host"`
	Username string      `json:"username"`
	Password string      `json:"password"`
	Queries  []WmiQuery  `json:"queries"`
	Methods  []WmiMethod `json:"methods"`
}

type WmiQuery struct {
	Namespace     string   `json:"namespace"`
	ClassName     string   `json:"class_name"`
	Properties    []string `json:"properties"`
	Filter        string   `json:"filter"`
	TagProperties []string `json:"tag_properties"`
}

type WmiMethod struct {
	Namespace     string                 `json:"namespace"`
	ClassName     string                 `json:"class_name"`
	Method        string                 `json:"method"`
	TagProperties []string               `json:"tag_properties"`
	Arguments     map[string]interface{} `json:"arguments"`
	Fields        map[string]string      `json:"fields"`
}
