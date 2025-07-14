package translations

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestBodyMappingError",
			Description: "Error",
			Other:       "Error mapping request body",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestQueryMappingError",
			Description: "Error",
			Other:       "Error mapping request query",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestURIMappingError",
			Description: "Error",
			Other:       "Error mapping URI parameters",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "QueryParamMissingError",
			Description: "Error",
			Other:       "A required query parameter {{.Name}} is missing",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "QueryParamInvalidError",
			Description: "Error",
			Other:       "Invalid query parameter {{.Name}} value, it can be one of {{.Values}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "BodyParamMissingError",
			Description: "Error",
			Other:       "Missing required parameter {{.Name}} from request body",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "BodyParamInvalidError",
			Description: "Error",
			Other:       "Incorrect parameter {{.Name}} value from request body, it can be one of {{.Values}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English), "ru").MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestValidationError",
			Description: "Error",
			Other:       "Error validating request body",
		},
	})
}
