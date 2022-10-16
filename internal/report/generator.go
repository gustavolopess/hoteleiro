package report

type Generator interface {
	GenerateShortReport(dateBegin, dateEnd string)
	GenerateLongReport(dateBegin, dateEnd string)
}