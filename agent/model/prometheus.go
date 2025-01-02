package model

type Prometheus struct {
	ID                      int                 `json:"id,omitempty"`
	URLs                    []string            `json:"urls,omitempty"`
	MetricVersion           int                 `json:"metric_version,omitempty"`
	URLTag                  string              `json:"url_tag,omitempty"`
	IgnoreTimestamp         bool                `json:"ignore_timestamp,omitempty"`
	ContentTypeOverride     string              `json:"content_type_override,omitempty"`
	KubernetesServices      []string            `json:"kubernetes_services,omitempty"`
	KubeConfig              string              `json:"kube_config,omitempty"`
	MonitorKubernetesPods   bool                `json:"monitor_kubernetes_pods,omitempty"`
	MonitorPodsMethod       string              `json:"monitor_kubernetes_pods_method,omitempty"`
	MonitorPodsScheme       string              `json:"monitor_kubernetes_pods_scheme,omitempty"`
	MonitorPodsPort         string              `json:"monitor_kubernetes_pods_port,omitempty"`
	MonitorPodsPath         string              `json:"monitor_kubernetes_pods_path,omitempty"`
	PodScrapeScope          string              `json:"pod_scrape_scope,omitempty"`
	NodeIP                  string              `json:"node_ip,omitempty"`
	PodScrapeInterval       int                 `json:"pod_scrape_interval,omitempty"`
	ContentLengthLimit      string              `json:"content_length_limit,omitempty"`
	MonitorPodsNamespace    string              `json:"monitor_kubernetes_pods_namespace,omitempty"`
	PodNamespaceLabelName   string              `json:"pod_namespace_label_name,omitempty"`
	KubernetesLabelSelector string              `json:"kubernetes_label_selector,omitempty"`
	KubernetesFieldSelector string              `json:"kubernetes_field_selector,omitempty"`
	PodAnnotationInclude    []string            `json:"pod_annotation_include,omitempty"`
	PodAnnotationExclude    []string            `json:"pod_annotation_exclude,omitempty"`
	PodLabelInclude         []string            `json:"pod_label_include,omitempty"`
	PodLabelExclude         []string            `json:"pod_label_exclude,omitempty"`
	CacheRefreshInterval    int                 `json:"cache_refresh_interval,omitempty"`
	Consul                  *PrometheusConsul   `json:"consul,omitempty"`
	BearerToken             string              `json:"bearer_token,omitempty"`
	BearerTokenString       string              `json:"bearer_token_string,omitempty"`
	Username                string              `json:"username,omitempty"`
	Password                string              `json:"password,omitempty"`
	HTTPHeaders             map[string]string   `json:"http_headers,omitempty"`
	Timeout                 string              `json:"timeout,omitempty"`
	ResponseTimeout         string              `json:"response_timeout,omitempty"`
	UseSystemProxy          bool                `json:"use_system_proxy,omitempty"`
	HTTPProxyURL            string              `json:"http_proxy_url,omitempty"`
	TLSConfig               *TLSConfig          `json:"tls_config,omitempty"`
	EnableRequestMetrics    bool                `json:"enable_request_metrics,omitempty"`
	NamespaceAnnotationPass map[string][]string `json:"namespace_annotation_pass,omitempty"`
	NamespaceAnnotationDrop map[string][]string `json:"namespace_annotation_drop,omitempty"`
}

type PrometheusConsul struct {
	Enabled       bool                    `json:"enabled,omitempty"`
	Agent         string                  `json:"agent,omitempty"`
	QueryInterval string                  `json:"query_interval,omitempty"`
	Queries       []PrometheusConsulQuery `json:"queries,omitempty"`
}

type PrometheusConsulQuery struct {
	Name string            `json:"name,omitempty"`
	Tag  string            `json:"tag,omitempty"`
	URL  string            `json:"url,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

type TLSConfig struct {
	CAFile              string `json:"tls_ca,omitempty"`
	CertFile            string `json:"tls_cert,omitempty"`
	KeyFile             string `json:"tls_key,omitempty"`
	InsecureSkipVerify  bool   `json:"insecure_skip_verify,omitempty"`
	ServerName          string `json:"tls_server_name,omitempty"`
	RenegotiationMethod string `json:"tls_renegotiation_method,omitempty"`
	Enable              bool   `json:"tls_enable,omitempty"`
}
