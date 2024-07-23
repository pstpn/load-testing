package worker

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"

	"sberhl/pkg/logger"
)

type Worker interface {
	Run() error
}

type workerImpl struct {
	l       logger.Interface
	url     string
	timeout float64
	threads int64
}

func NewWorker(url string, timeout time.Duration, threads int64, l logger.Interface) Worker {
	return &workerImpl{
		l:       l,
		url:     url,
		timeout: float64(timeout.Milliseconds()),
		threads: threads,
	}
}

func (w *workerImpl) runThread(errs chan error) {
	page, err := w.connectToPage()
	if err != nil {
		w.l.Errorf("failed to connect to web page: %s", err.Error())
		errs <- fmt.Errorf("connect to web page: %w", err)
		return
	}

	err = page.Locator(button).Click()
	if err != nil {
		w.l.Errorf("failed to click on start button: %s", err.Error())
		errs <- fmt.Errorf("click on start button: %w", err)
		return
	}

	var stop bool
	for stop, err = w.reloadAndCheck(page); err == nil && !stop; stop, err = w.reloadAndCheck(page) {
		err := w.processPage(page)
		if err != nil {
			w.l.Errorf("failed to process page: %s", err.Error())
			errs <- fmt.Errorf("process page: %w", err)
			return
		}

		err = page.Locator(rootHTML + button).Click()
		if err != nil {
			w.l.Errorf("failed to click submit button: %s", err.Error())
			errs <- fmt.Errorf("click submit button: %w", err)
			return
		}
	}
	if err != nil {
		w.l.Errorf("failed to reload page and check: %s", err.Error())
		errs <- fmt.Errorf("reload page and check: %w", err)
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

	return false, nil
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
			w.l.Infof("run thread with number %d", i+1)
			defer wg.Done()
			w.runThread(errs)
			w.l.Infof("thread with number %d finished successfully", i+1)
		}()
	}

	wg.Wait()
	close(errs)

	return <-firstErr
}
