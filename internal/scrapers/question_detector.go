package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

// CustomQuestion represents a detected screening question
type CustomQuestion struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"` // textarea, select, radio, checkbox
	Label    string   `json:"label"`
	Required bool     `json:"required"`
	Selector string   `json:"selector"`
	Options  []string `json:"options,omitempty"` // For select/radio/checkbox
}

// DetectCustomQuestions finds non-standard form fields that need user input
func DetectCustomQuestions(ctx context.Context) ([]CustomQuestion, error) {
	var questions []CustomQuestion

	// JavaScript to detect custom questions
	jsCode := `
	(function() {
		// Standard field patterns to SKIP
		const standardPatterns = [
			/first.*name/i,
			/last.*name/i,
			/^name$/i,
			/email/i,
			/phone/i,
			/telephone/i,
			/mobile/i,
			/address/i,
			/street/i,
			/city/i,
			/state/i,
			/zip/i,
			/postal/i,
			/resume/i,
			/cv/i,
			/linkedin/i,
			/portfolio/i,
			/website/i,
		];

		function isStandardField(label) {
			if (!label) return false;
			const lowerLabel = label.toLowerCase().trim();
			return standardPatterns.some(pattern => pattern.test(lowerLabel));
		}

		function getLabel(element) {
			// Try label element
			const labelElement = element.closest('label') ||
			                    document.querySelector('label[for="' + element.id + '"]');
			if (labelElement) {
				return labelElement.textContent.trim();
			}

			// Try aria-label
			if (element.getAttribute('aria-label')) {
				return element.getAttribute('aria-label').trim();
			}

			// Try placeholder
			if (element.placeholder) {
				return element.placeholder.trim();
			}

			// Try name attribute
			if (element.name) {
				return element.name.replace(/[-_]/g, ' ').trim();
			}

			return '';
		}

		function isRequired(element) {
			return element.hasAttribute('required') ||
			       element.getAttribute('aria-required') === 'true' ||
			       element.closest('.required') !== null;
		}

		function getSelector(element) {
			// Try ID first
			if (element.id) {
				return '#' + element.id;
			}

			// Try name
			if (element.name) {
				return element.tagName.toLowerCase() + '[name="' + element.name + '"]';
			}

			// Fallback to index
			const siblings = Array.from(element.parentElement.querySelectorAll(element.tagName));
			const index = siblings.indexOf(element);
			return element.tagName.toLowerCase() + ':nth-of-type(' + (index + 1) + ')';
		}

		const customQuestions = [];

		// Find textareas
		const textareas = document.querySelectorAll('textarea');
		textareas.forEach((textarea, index) => {
			const label = getLabel(textarea);
			if (!isStandardField(label) && textarea.offsetParent !== null) {
				customQuestions.push({
					id: 'textarea_' + index,
					type: 'textarea',
					label: label || 'Text Response ' + (index + 1),
					required: isRequired(textarea),
					selector: getSelector(textarea),
					options: []
				});
			}
		});

		// Find select dropdowns
		const selects = document.querySelectorAll('select');
		selects.forEach((select, index) => {
			const label = getLabel(select);
			if (!isStandardField(label) && select.offsetParent !== null) {
				const options = Array.from(select.options)
					.map(opt => opt.text.trim())
					.filter(text => text && text !== '' && text !== 'Select...' && text !== 'Choose...');

				customQuestions.push({
					id: 'select_' + index,
					type: 'select',
					label: label || 'Selection ' + (index + 1),
					required: isRequired(select),
					selector: getSelector(select),
					options: options
				});
			}
		});

		// Find radio button groups
		const radioGroups = {};
		document.querySelectorAll('input[type="radio"]').forEach(radio => {
			if (!radio.name || radio.offsetParent === null) return;

			if (!radioGroups[radio.name]) {
				const label = getLabel(radio.closest('fieldset') || radio.parentElement);
				if (!isStandardField(label)) {
					radioGroups[radio.name] = {
						id: 'radio_' + radio.name,
						type: 'radio',
						label: label || radio.name,
						required: isRequired(radio),
						selector: 'input[name="' + radio.name + '"]',
						options: []
					};
				}
			}

			if (radioGroups[radio.name]) {
				const optionLabel = getLabel(radio);
				if (optionLabel) {
					radioGroups[radio.name].options.push(optionLabel);
				}
			}
		});

		Object.values(radioGroups).forEach(group => {
			customQuestions.push(group);
		});

		return customQuestions;
	})();
	`

	// Execute JavaScript and get results
	err := chromedp.Run(ctx,
		chromedp.Evaluate(jsCode, &questions),
	)

	if err != nil {
		log.Printf("Error detecting custom questions: %v", err)
		return nil, fmt.Errorf("failed to detect questions: %w", err)
	}

	log.Printf("Detected %d custom questions", len(questions))
	for i, q := range questions {
		log.Printf("  Question %d: %s (%s) - Required: %v", i+1, q.Label, q.Type, q.Required)
	}

	return questions, nil
}

// HasNextButton checks if there's a "Next" or "Continue" button for multi-page forms
func HasNextButton(ctx context.Context) (bool, error) {
	var hasButton bool

	jsCode := `
	(function() {
		const buttons = Array.from(document.querySelectorAll('button, input[type="button"], input[type="submit"]'));
		return buttons.some(btn => {
			const text = (btn.textContent || btn.value || '').toLowerCase();
			return text.includes('next') || text.includes('continue');
		});
	})();
	`

	err := chromedp.Run(ctx,
		chromedp.Evaluate(jsCode, &hasButton),
	)

	if err != nil {
		return false, fmt.Errorf("failed to check for next button: %w", err)
	}

	return hasButton, nil
}

// ClickNextButton clicks the "Next" or "Continue" button
func ClickNextButton(ctx context.Context) error {
	// Try multiple selectors for next button
	nextSelectors := []string{
		`button:contains("Next")`,
		`button:contains("Continue")`,
		`input[value*="Next"]`,
		`input[value*="Continue"]`,
		`button[id*="next"]`,
		`button[id*="continue"]`,
	}

	for _, selector := range nextSelectors {
		err := chromedp.Run(ctx,
			chromedp.Click(selector, chromedp.NodeVisible),
		)
		if err == nil {
			log.Printf("Clicked next button with selector: %s", selector)
			return nil
		}
	}

	return fmt.Errorf("could not find or click next button")
}

// FillAnswer fills in a custom question answer
func FillAnswer(ctx context.Context, question CustomQuestion, answer string) error {
	switch question.Type {
	case "textarea", "text":
		return chromedp.Run(ctx,
			chromedp.SendKeys(question.Selector, answer, chromedp.NodeVisible),
		)

	case "select":
		// Find the option that matches the answer
		jsCode := fmt.Sprintf(`
			(function() {
				const select = document.querySelector('%s');
				if (!select) return false;

				const options = Array.from(select.options);
				const matchingOption = options.find(opt =>
					opt.text.toLowerCase().includes('%s') ||
					opt.value.toLowerCase().includes('%s')
				);

				if (matchingOption) {
					select.value = matchingOption.value;
					select.dispatchEvent(new Event('change', { bubbles: true }));
					return true;
				}
				return false;
			})();
		`, question.Selector, answer, answer)

		var success bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(jsCode, &success),
		)
		if err != nil || !success {
			return fmt.Errorf("failed to set select value")
		}
		return nil

	case "radio":
		// Find and click the radio button that matches the answer
		jsCode := fmt.Sprintf(`
			(function() {
				const radios = document.querySelectorAll('%s');
				for (let radio of radios) {
					const label = radio.closest('label') || document.querySelector('label[for="' + radio.id + '"]');
					const labelText = label ? label.textContent.toLowerCase() : '';
					if (labelText.includes('%s') || radio.value.toLowerCase().includes('%s')) {
						radio.checked = true;
						radio.dispatchEvent(new Event('change', { bubbles: true }));
						return true;
					}
				}
				return false;
			})();
		`, question.Selector, answer, answer)

		var success bool
		err := chromedp.Run(ctx,
			chromedp.Evaluate(jsCode, &success),
		)
		if err != nil || !success {
			return fmt.Errorf("failed to select radio option")
		}
		return nil

	default:
		return fmt.Errorf("unsupported question type: %s", question.Type)
	}
}

// QuestionToJSON converts questions to JSON for database storage
func QuestionToJSON(questions []CustomQuestion) []byte {
	data, _ := json.Marshal(questions)
	return data
}

// JSONToQuestions converts JSON from database back to question structs
func JSONToQuestions(data []byte) ([]CustomQuestion, error) {
	var questions []CustomQuestion
	if len(data) == 0 {
		return questions, nil
	}
	err := json.Unmarshal(data, &questions)
	return questions, err
}
