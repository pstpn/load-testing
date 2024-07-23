package worker

import "github.com/playwright-community/playwright-go"

const (
	rootHTML = `html body form p `
	bodyHTML = `html body`

	inputText    = `input[type=text]`
	inputRadio   = `input[type=radio]`
	selectOption = `select option`

	inputNameFmt  = `input[name='%s']`
	inputValueFmt = `input[value='%s']`
	selectNameFmt = `select[name='%s']`

	button = `button`

	passedMsg         = "Test successfully passed"
	toManyRequestsMsg = "too many requests"
)

func (w *workerImpl) maxLengthRadio(radioAttr []playwright.Locator) (map[string]string, error) {
	radioResult := make(map[string]string)

	for _, radio := range radioAttr {
		name, err := radio.GetAttribute("name")
		if err != nil {
			return nil, err
		}
		value, err := radio.GetAttribute("value")
		if err != nil {
			return nil, err
		}

		if res, ok := radioResult[name]; !ok || len(res) < len(value) {
			radioResult[name] = value
		}
	}

	return radioResult, nil
}

func (w *workerImpl) maxLengthSelect(selectAttr []playwright.Locator) (map[string]string, error) {
	selectResult := make(map[string]string)

	for _, sel := range selectAttr {
		name, err := sel.Locator("..").GetAttribute("name")
		if err != nil {
			return nil, err
		}
		value, err := sel.GetAttribute("value")
		if err != nil {
			return nil, err
		}

		if res, ok := selectResult[name]; !ok || len(res) < len(value) {
			selectResult[name] = value
		}
	}

	return selectResult, nil
}
