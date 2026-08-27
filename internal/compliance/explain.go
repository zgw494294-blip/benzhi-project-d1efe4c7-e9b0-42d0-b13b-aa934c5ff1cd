package compliance

func RuleCount(version string) int {
	if version == "GB-2024" {
		return 6
	}
	return 0
}
