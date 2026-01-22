package indices

type TokenizerConfig struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

type AnalyzerConfig struct {
	Type      string `json:"type"`
	Tokenizer string `json:"tokenizer"`
}

type AnalysisConfig struct {
	Tokenizers map[string]TokenizerConfig `json:"tokenizer"`
	Analyzers  map[string]AnalyzerConfig  `json:"analyzer"`
}

type IndexRequest struct {
	InstanceID       uint           `json:"instance_id" binding:"required"`      // 集群实例 ID，必填（使用uint匹配UNSIGNED INT）
	IndexName        string         `json:"index_name" binding:"required"`       // 索引名称，必填
	NumberOfShards   int            `json:"number_of_shards" binding:"required"` // 分片数量，必填
	NumberOfReplicas int            `json:"number_of_replicas"`                  // 副本数量，非必填
	RefreshInterval  string         `json:"refresh_interval"`                    // 刷新间隔，非必填
	Alias            string         `json:"alias"`                               // 别名，非必填
	Analysis         AnalysisConfig `json:"analysis"`                            // 分词器配置，非必填
}

type IndexDeleteRequest struct {
	InstanceID uint     `json:"instance_id"` // 使用uint匹配UNSIGNED INT
	IndexNames []string `json:"index_names"`
}

type IndexUpdateRequest struct {
	InstanceID       uint    `json:"instance_id" binding:"required"` // 使用uint匹配UNSIGNED INT
	IndexName        string  `json:"index_name" binding:"required"`
	NumberOfReplicas *int    `json:"number_of_replicas"`
	RefreshInterval  *string `json:"refresh_interval"`
}
