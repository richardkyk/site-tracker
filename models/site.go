package models

type Site struct {
	ID          string `dynamodbav:"id"`
	URL         string `dynamodbav:"url"`
	Selector    string `dynamodbav:"selector"`
	Regex       string `dynamodbav:"regex"`
	Expected    string `dynamodbav:"expected"`
	Email       string `dynamodbav:"email"`
	ShouldCheck bool   `dynamodbav:"shouldCheck"`
}
