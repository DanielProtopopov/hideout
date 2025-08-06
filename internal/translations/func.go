package translations

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestBodyMappingError",
			Description: "Error",
			Other:       "Error mapping request body",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestQueryMappingError",
			Description: "Error",
			Other:       "Error mapping request query",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestURIMappingError",
			Description: "Error",
			Other:       "Error mapping URI parameters",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "QueryParamMissingError",
			Description: "Error",
			Other:       "A required query parameter {{.Name}} is missing",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "QueryParamInvalidError",
			Description: "Error",
			Other:       "Invalid query parameter {{.Name}} value, it can be one of {{.Values}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "BodyParamMissingError",
			Description: "Error",
			Other:       "Missing required parameter {{.Name}} from request body",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "BodyParamInvalidError",
			Description: "Error",
			Other:       "Incorrect parameter {{.Name}} value from request body, it can be one of {{.Values}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RequestValidationError",
			Description: "Error",
			Other:       "Error validating request body",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "CopySecretsFoldersError",
			Description: "Error",
			Other:       "Error copying secrets/folders to folder with UID of {{.UID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "CreateSecretsServiceError",
			Description: "Error",
			Other:       "Error creating secrets service",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "DeleteFolderError",
			Description: "Error",
			Other:       "Error deleting folder with UID of {{.UID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "DeleteSecretError",
			Description: "Error",
			Other:       "Error deleting secret with UID of {{.UID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "FolderNotFoundError",
			Description: "Error",
			Other:       "Error finding folder with UID of {{.UID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "GetSecretFolderError",
			Description: "Error",
			Other:       "Error retrieving folder of copied secret with UID of {{.UID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "GetFolderByUIDError",
			Description: "Error",
			Other:       "Error retrieving folder with UID of {{.UID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "GetFoldersError",
			Description: "Error",
			Other:       "Error retrieving folders",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "GetSecretByUIDError",
			Description: "Error",
			Other:       "Error retrieving secret with UID of {{.UID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "GetSecretsError",
			Description: "Error",
			Other:       "Error retrieving secrets",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "OrderError",
			Description: "Error",
			Other:       "Error validating ordering",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "PaginationError",
			Description: "Error",
			Other:       "Error validating pagination",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "SecretAlreadyExistsError",
			Description: "Error",
			Other:       "Secret with name of \"{{.Name}}\" already exists in folder with UID of {{.FolderUID}}",
		},
	})
}

func _() string {
	return i18n.NewLocalizer(i18n.NewBundle(language.English)).MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "UpdateSecretError",
			Description: "Error",
			Other:       "Error updating secret with UID of {{.UID}}",
		},
	})
}
