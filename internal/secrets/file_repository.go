package secrets

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"github.com/gocarina/gocsv"
	"github.com/pkg/errors"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"hideout/internal/pkg/extra"
	"os"
)

type FileRepository struct {
	Filename           string
	EncodingType       uint
	inMemoryRepository *InMemoryRepository
}

func NewFileRepository(filename string, encodingType uint, inMemoryRep *InMemoryRepository) FileRepository {
	return FileRepository{Filename: filename, EncodingType: encodingType, inMemoryRepository: inMemoryRep}
}

func (m FileRepository) GetID(ctx context.Context) (uint, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetID(ctx)
	}

	secrets, errLoadSecrets := m.Load(ctx)
	if errLoadSecrets != nil {
		return 0, errors.Wrap(errLoadSecrets, "Failed to load secrets from File")
	}

	var maxID = uint(0)
	for _, secret := range secrets {
		if secret.ID >= maxID {
			maxID = secret.ID
		}
	}

	return maxID, nil
}

func (m FileRepository) Load(ctx context.Context) ([]Secret, error) {
	var secrets []Secret
	errDecode := m.decode(&secrets)
	return secrets, errDecode
}

func (m FileRepository) GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[uint]*Secret)
	for _, result := range results {
		mapResults[result.ID] = result
	}

	return mapResults, nil
}

func (m FileRepository) GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByUID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[string]*Secret)
	for _, result := range results {
		mapResults[result.UID] = result
	}

	return mapResults, nil
}

func (m FileRepository) Get(ctx context.Context, params ListSecretParams) ([]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	results, errLoadSecrets := m.Load(ctx)
	if errLoadSecrets != nil {
		return nil, errLoadSecrets
	}

	inMemoryRepository := NewInMemoryRepository(&results)
	return inMemoryRepository.Get(ctx, params)
}

func (m FileRepository) GetMapByFolder(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByFolder(ctx, params)
	}

	return nil, apperror.ErrNotImplemented
}

func (m FileRepository) GetByID(ctx context.Context, id uint) (*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	results, errGetResults := m.GetMapByID(ctx, ListSecretParams{
		ListParams: generics.ListParams{IDs: []uint{id}, Deleted: model.No},
	})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := results[id]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m FileRepository) GetByUID(ctx context.Context, uid string) (*Secret, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	results, errGetResults := m.GetMapByUID(ctx, ListSecretParams{
		ListParams: generics.ListParams{UIDs: []string{uid}, Deleted: model.No},
	})
	if errGetResults != nil {
		return nil, errGetResults
	}

	result, exists := results[uid]
	if !exists {
		return nil, apperror.ErrRecordNotFound
	}

	return result, nil
}

func (m FileRepository) Update(ctx context.Context, secret Secret) (*Secret, error) {
	var inMemoryRepository = m.inMemoryRepository
	if inMemoryRepository == nil {
		secrets, errLoadSecrets := m.Load(ctx)
		if errLoadSecrets != nil {
			return nil, errLoadSecrets
		}
		inMemoryRepository = NewInMemoryRepository(&secrets)
	}

	updatedSecret, errUpdateSecret := inMemoryRepository.Update(ctx, secret)
	if errUpdateSecret != nil {
		return nil, errUpdateSecret
	}
	secretPtrs, errGetSecrets := inMemoryRepository.Get(ctx, ListSecretParams{ListParams: generics.ListParams{Deleted: model.YesOrNo}})
	if errGetSecrets != nil {
		return nil, errGetSecrets
	}
	var secrets []Secret
	for _, secretPtr := range secretPtrs {
		secrets = append(secrets, *secretPtr)
	}
	return updatedSecret, m.encode(&secrets)
}

func (m FileRepository) Create(ctx context.Context, secret Secret) (*Secret, error) {
	if m.inMemoryRepository != nil {
		createdSecret, errCreateSecret := m.inMemoryRepository.Create(ctx, secret)
		if errCreateSecret != nil {
			return nil, errors.Wrapf(errCreateSecret, "Error creating secret with ID of %d in memory", secret.ID)
		}

		secretPtrs, errGetSecrets := m.inMemoryRepository.Get(ctx, ListSecretParams{ListParams: generics.ListParams{Deleted: model.YesOrNo}})
		if errGetSecrets != nil {
			return nil, errGetSecrets
		}
		var secrets []Secret
		for _, secretPtr := range secretPtrs {
			secrets = append(secrets, *secretPtr)
		}
		errEncode := m.encode(&secrets)
		return createdSecret, errEncode
	}

	// Done this way because file may have a duplicate entry and needs to be
	// loaded to check
	var secrets []Secret
	errDecode := m.decode(&secrets)
	if errDecode != nil {
		return nil, errDecode
	}

	inMemoryRepository := NewInMemoryRepository(&secrets)
	createdSecret, errCreateSecret := inMemoryRepository.Create(ctx, secret)
	if errCreateSecret != nil {
		return nil, errors.Wrapf(errCreateSecret, "Error creating secret with ID of %d in memory", secret.ID)
	}

	errEncode := m.encode(&secrets)
	if errEncode != nil {
		return nil, errEncode
	}

	return createdSecret, nil
}

func (m FileRepository) Count(ctx context.Context, params ListSecretParams) (uint, error) {
	// These are not needed when performing filtering and counting
	params.Pagination = pagination.Pagination{PerPage: 0, Page: 0}
	params.Order = []ordering.Order{}

	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Count(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return 0, errGetResults
	}

	return uint(len(results)), nil
}

func (m FileRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	var inMemoryRepository = m.inMemoryRepository
	if inMemoryRepository == nil {
		secrets, errLoadSecrets := m.Load(ctx)
		if errLoadSecrets != nil {
			return errLoadSecrets
		}
		inMemoryRepository = NewInMemoryRepository(&secrets)
	}

	errDelete := inMemoryRepository.Delete(ctx, id, forceDelete)
	if errDelete != nil {
		return errDelete
	}

	secretPtrs, errGetSecrets := inMemoryRepository.Get(ctx, ListSecretParams{ListParams: generics.ListParams{Deleted: model.YesOrNo}})
	if errGetSecrets != nil {
		return errGetSecrets
	}
	var secrets []Secret
	for _, secretPtr := range secretPtrs {
		secrets = append(secrets, *secretPtr)
	}
	return m.encode(&secrets)
}

func (m FileRepository) encode(data *[]Secret) error {
	fileWriter, errOpenFile := os.OpenFile(m.Filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if errOpenFile != nil {
		return errOpenFile
	}
	defer fileWriter.Close()
	switch m.EncodingType {
	case extra.Encoding_JSON:
		{
			encoder := json.NewEncoder(fileWriter)
			encoder.SetIndent("", " ")
			return encoder.Encode(data)
		}
	case extra.Encoding_CSV:
		{
			return gocsv.UnmarshalFile(fileWriter, data)
		}
	case extra.Encoding_GOB:
		{
			encoder := gob.NewEncoder(fileWriter)
			return encoder.Encode(data)
		}
	case extra.Encoding_XML:
		{
			xmlData, errMarshal := xml.Marshal(data)
			if errMarshal != nil {
				return errMarshal
			}
			_, errWrite := fileWriter.Write(xmlData)
			return errWrite
		}
	}

	return apperror.ErrNotImplemented
}

func (m FileRepository) decode(data *[]Secret) error {
	_, errFileExists := os.Stat(m.Filename)
	if errors.Is(errFileExists, os.ErrNotExist) {
		return nil
	}

	fileReader, errOpenFile := os.OpenFile(m.Filename, os.O_RDONLY, 0644)
	if errOpenFile != nil {
		return errOpenFile
	}
	defer fileReader.Close()

	switch m.EncodingType {
	case extra.Encoding_JSON:
		{
			decoder := json.NewDecoder(fileReader)
			return decoder.Decode(data)
		}
	case extra.Encoding_CSV:
		{
			return gocsv.MarshalFile(data, fileReader)
		}
	case extra.Encoding_GOB:
		{
			decoder := gob.NewDecoder(fileReader)
			return decoder.Decode(data)
		}
	case extra.Encoding_XML:
		{
			buf := new(bytes.Buffer)
			_, errRead := buf.ReadFrom(fileReader)
			if errRead != nil {
				return errRead
			}

			return xml.Unmarshal(buf.Bytes(), data)
		}
	}

	return apperror.ErrNotImplemented
}
