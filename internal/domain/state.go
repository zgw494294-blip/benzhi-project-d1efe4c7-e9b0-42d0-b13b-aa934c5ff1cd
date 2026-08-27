package domain

func Allowed(from, to Status) bool {
	switch from {
	case Draft:
		return to == Checked || to == ReworkRequired
	case Checked:
		return to == ReworkRequired || to == ReadyForReview
	case ReworkRequired:
		return to == Checked
	case ReadyForReview:
		return to == Approved || to == ReworkRequired
	case Approved:
		return to == Frozen
	case Frozen:
		return to == Authorized
	default:
		return false
	}
}
