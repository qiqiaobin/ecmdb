package web

type RelationType struct {
	ID             int64
	Name           string
	UID            string
	SourceDescribe string
	TargetDescribe string
}

type ModelRelation struct {
	ID              int64  `json:"id"`
	SourceModelUID  string `json:"source_model_uid"`
	TargetModelUID  string `json:"target_model_uid"`
	RelationTypeUID string `json:"relation_type_uid"`
	RelationName    string `json:"relation_name"`
	Mapping         string `json:"mapping"`
}
