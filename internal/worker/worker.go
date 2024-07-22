package worker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

type Worker interface {
	Run() error
}

type workerImpl struct {
	url     string
	timeout float64
	threads int64
}

func NewWorker(url string, timeout time.Duration, threads int64) Worker {
	return &workerImpl{
		url:     url,
		timeout: float64(timeout.Milliseconds()),
		threads: threads,
	}
}

func (w *workerImpl) Run() error {
	errs := make(chan error)
	firstErr := make(chan error)
	defer close(firstErr)

	wg := &sync.WaitGroup{}

	go func() {
		for err := range errs {
			firstErr <- err
		}
		firstErr <- nil
	}()

	for i := range w.threads {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.runThread(int(i), errs)
		}()
	}

	wg.Wait()
	close(errs)

	return <-firstErr
}

func (w *workerImpl) runThread(i int, errs chan error) {
	page, err := w.connectToPage()
	if err != nil {
		errs <- fmt.Errorf("connect to web page: %w", err)
		return
	}

	err = page.Locator(button).Click()
	if err != nil {
		errs <- fmt.Errorf("click on start button: %w", err)
		return
	}

	var stop bool
	for stop, err = w.reloadAndCheck(page); err == nil && !stop; stop, err = w.reloadAndCheck(page) {
		err := w.processPage(page)
		if err != nil {
			errs <- fmt.Errorf("process page: %w", err)
			return
		}

		err = page.Locator(rootHTML + button).Click()
		if err != nil {
			errs <- fmt.Errorf("click submit button: %w", err)
			return
		}
	}
	if err != nil {
		errs <- fmt.Errorf("inner HTML: %w", err)
	}
}

func (w *workerImpl) connectToPage() (playwright.Page, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return nil, err
	}

	page, err := browser.NewPage()
	if err != nil {
		return nil, err
	}

	if _, err = page.Goto(w.url); err != nil {
		return nil, err
	}
	page.SetDefaultTimeout(w.timeout)

	return page, nil
}

func (w *workerImpl) reloadAndCheck(page playwright.Page) (bool, error) {
	var html string
	var err error

	for html, err = page.Locator(bodyHTML).InnerHTML(); err == nil && strings.Contains(html, toManyRequestsMsg); html, err = page.Locator(bodyHTML).InnerHTML() {
		_, err = page.Reload()
		if err != nil {
			return false, err
		}
	}
	if err != nil {
		return false, err
	}

	if strings.Contains(html, passedMsg) {
		return true, nil
	}

	//screenshot, err = page.Locator(`html`).Screenshot()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = os.WriteFile(fmt.Sprintf("temp%d_%d.png", i, j), screenshot, 0777)
	//if err != nil {
	//	log.Fatal(err)
	//}

	return false, nil
}

func (w *workerImpl) processPage(page playwright.Page) error {
	err := w.insertText(page)
	if err != nil {
		return fmt.Errorf("insert text to fields: %w", err)
	}
	err = w.clickRadio(page)
	if err != nil {
		return fmt.Errorf("click to radio fields: %w", err)
	}
	err = w.chooseSelect(page)
	if err != nil {
		return fmt.Errorf("choose select field: %w", err)
	}

	return nil
}

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

func (w *workerImpl) insertText(page playwright.Page) error {
	allText, err := page.Locator(rootHTML + inputText).All()
	if err != nil {
		return err
	}

	for _, loc := range allText {
		name, err := loc.GetAttribute("name")
		if err != nil {
			return err
		}
		err = page.Locator(fmt.Sprintf(rootHTML+inputNameFmt, name)).Fill("test")
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *workerImpl) clickRadio(page playwright.Page) error {
	allRadio, err := page.Locator(rootHTML + inputRadio).All()
	if err != nil {
		return err
	}

	maxRadio, err := w.maxLengthRadio(allRadio)
	if err != nil {
		return err
	}

	for _, value := range maxRadio {
		err = page.Locator(fmt.Sprintf(rootHTML+inputValueFmt, value)).Click()
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *workerImpl) chooseSelect(page playwright.Page) error {
	allSelect, err := page.Locator(rootHTML + selectOption).All()
	if err != nil {
		return err
	}

	maxSelect, err := w.maxLengthSelect(allSelect)
	if err != nil {
		return err
	}

	for name, value := range maxSelect {
		err = page.Locator(fmt.Sprintf(rootHTML+selectNameFmt, name)).Click()
		if err != nil {
			return err
		}

		_, err = page.Locator(fmt.Sprintf(rootHTML+selectNameFmt, name)).SelectOption(playwright.SelectOptionValues{
			Labels: &[]string{value},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
