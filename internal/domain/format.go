package domain

import "fmt"

func FormatMillimeters(v float64) string { return fmt.Sprintf("%.3fmm", v) }
func FormatRatio(v float64) string       { return fmt.Sprintf("%.3f", v) }
func StatusLabel(s Status) string {
	switch s {
	case Draft:
		return "草稿"
	case Checked:
		return "已校核"
	case ReworkRequired:
		return "需要整改"
	case ReadyForReview:
		return "待复核"
	case Approved:
		return "已批准"
	case Frozen:
		return "已冻结"
	case Authorized:
		return "已授权"
	default:
		return string(s)
	}
}
func SeverityLabel(s Severity) string {
	switch s {
	case Pass:
		return "通过"
	case Warn:
		return "警告"
	case Block:
		return "阻断"
	default:
		return string(s)
	}
}
