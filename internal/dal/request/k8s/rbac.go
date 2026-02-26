package k8s

// ==================== Role ====================

// RoleListItem Role列表项
type RoleListItem struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	RulesCount int    `json:"rulesCount"`
	Age        int64  `json:"age"`
}

// PolicyRule 权限规则
type PolicyRule struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// RoleCreateRequest 创建Role请求
type RoleCreateRequest struct {
	Name      string       `json:"name" binding:"required"`
	Namespace string       `json:"namespace" binding:"required"`
	Rules     []PolicyRule `json:"rules"`
	YAML      string       `json:"yaml"`
}

// RoleUpdateRequest 更新Role请求
type RoleUpdateRequest struct {
	Rules []PolicyRule `json:"rules"`
	YAML  string       `json:"yaml"`
}

// ==================== ClusterRole ====================

// ClusterRoleListItem ClusterRole列表项
type ClusterRoleListItem struct {
	Name            string `json:"name"`
	RulesCount      int    `json:"rulesCount"`
	AggregationRule bool   `json:"aggregationRule"`
	Age             int64  `json:"age"`
}

// ClusterRoleCreateRequest 创建ClusterRole请求
type ClusterRoleCreateRequest struct {
	Name  string       `json:"name" binding:"required"`
	Rules []PolicyRule `json:"rules"`
	YAML  string       `json:"yaml"`
}

// ClusterRoleUpdateRequest 更新ClusterRole请求
type ClusterRoleUpdateRequest struct {
	Rules []PolicyRule `json:"rules"`
	YAML  string       `json:"yaml"`
}

// ==================== RoleBinding ====================

// RoleBindingSubject 绑定主体
type RoleBindingSubject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// RoleRef 角色引用
type RoleRef struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	APIGroup string `json:"apiGroup"`
}

// RoleBindingListItem RoleBinding列表项
type RoleBindingListItem struct {
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	RoleRefKind   string `json:"roleRefKind"`
	RoleRefName   string `json:"roleRefName"`
	SubjectsCount int    `json:"subjectsCount"`
	Age           int64  `json:"age"`
}

// RoleBindingCreateRequest 创建RoleBinding请求
type RoleBindingCreateRequest struct {
	Name      string               `json:"name" binding:"required"`
	Namespace string               `json:"namespace" binding:"required"`
	RoleRef   RoleRef              `json:"roleRef"`
	Subjects  []RoleBindingSubject `json:"subjects"`
	YAML      string               `json:"yaml"`
}

// RoleBindingUpdateRequest 更新RoleBinding请求
type RoleBindingUpdateRequest struct {
	Subjects []RoleBindingSubject `json:"subjects"`
	YAML     string               `json:"yaml"`
}

// ==================== ClusterRoleBinding ====================

// ClusterRoleBindingListItem ClusterRoleBinding列表项
type ClusterRoleBindingListItem struct {
	Name          string `json:"name"`
	RoleRefName   string `json:"roleRefName"`
	SubjectsCount int    `json:"subjectsCount"`
	Age           int64  `json:"age"`
}

// ClusterRoleBindingCreateRequest 创建ClusterRoleBinding请求
type ClusterRoleBindingCreateRequest struct {
	Name     string               `json:"name" binding:"required"`
	RoleRef  RoleRef              `json:"roleRef"`
	Subjects []RoleBindingSubject `json:"subjects"`
	YAML     string               `json:"yaml"`
}

// ClusterRoleBindingUpdateRequest 更新ClusterRoleBinding请求
type ClusterRoleBindingUpdateRequest struct {
	Subjects []RoleBindingSubject `json:"subjects"`
	YAML     string               `json:"yaml"`
}
