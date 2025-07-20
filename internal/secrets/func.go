package secrets

import (
	"slices"
)

func (params ListSecretParams) Apply(data map[string][]*Secret) (results map[string][]*Secret) {
	if len(params.IDs) != 0 {
		idResults := make(map[string][]*Secret)
		for pathVal, secretsEntry := range data {
			for _, secret := range secretsEntry {
				if slices.Index(params.IDs, secret.ID) != -1 {
					idResults[pathVal] = append(idResults[pathVal], secret)
				}
			}
		}
		results = idResults
	} else {
		results = data
	}

	if len(params.UIDs) != 0 {
		uidResults := make(map[string][]*Secret)
		for pathVal, secretsEntry := range data {
			for _, secret := range secretsEntry {
				if slices.Index(params.UIDs, secret.UID) != -1 {
					uidResults[pathVal] = append(uidResults[pathVal], secret)
				}
			}
		}
		results = uidResults
	}

	return results
}

func (params ListSecretParams) ApplyOrder(data map[string][]*Secret) (results map[string][]*Secret) {
	// @TODO Implement sorting
	return data
}
