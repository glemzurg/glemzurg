package model_data_type

func t_IntPtr(i int) *int {
	return &i
}

func t_BoolPtr(b bool) *bool {
	return &b
}

func t_StrPtr(s string) *string {
	return &s
}
