package folders

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

	folders, errLoadFolders := m.Load(ctx)
	if errLoadFolders != nil {
		return 0, errors.Wrap(errLoadFolders, "Failed to load folders from File")
	}

	var maxID = uint(0)
	for _, folder := range folders {
		if folder.ID >= maxID {
			maxID = folder.ID
		}
	}

	return maxID + 1, nil
}

func (m FileRepository) Load(ctx context.Context) ([]Folder, error) {
	var folders []Folder
	errDecode := m.decode(&folders)
	return folders, errDecode
}

func (m FileRepository) GetMapByID(ctx context.Context, params ListFolderParams) (map[uint]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[uint]*Folder)
	for _, result := range results {
		mapResults[result.ID] = result
	}

	return mapResults, nil
}

func (m FileRepository) GetMapByUID(ctx context.Context, params ListFolderParams) (map[string]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetMapByUID(ctx, params)
	}

	results, errGetResults := m.Get(ctx, params)
	if errGetResults != nil {
		return nil, errGetResults
	}

	var mapResults = make(map[string]*Folder)
	for _, result := range results {
		mapResults[result.UID] = result
	}

	return mapResults, nil
}

func (m FileRepository) Get(ctx context.Context, params ListFolderParams) ([]*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.Get(ctx, params)
	}

	results, errLoadFolders := m.Load(ctx)
	if errLoadFolders != nil {
		return nil, errLoadFolders
	}

	inMemoryRepository := NewInMemoryRepository(&results)
	return inMemoryRepository.Get(ctx, params)
}

func (m FileRepository) GetByID(ctx context.Context, id uint) (*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByID(ctx, id)
	}

	results, errGetResults := m.GetMapByID(ctx, ListFolderParams{
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

func (m FileRepository) GetByUID(ctx context.Context, uid string) (*Folder, error) {
	if m.inMemoryRepository != nil {
		return m.inMemoryRepository.GetByUID(ctx, uid)
	}

	results, errGetResults := m.GetMapByUID(ctx, ListFolderParams{
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

func (m FileRepository) Update(ctx context.Context, folder Folder) (*Folder, error) {
	var inMemoryRepository = m.inMemoryRepository
	if inMemoryRepository == nil {
		folders, errLoadFolders := m.Load(ctx)
		if errLoadFolders != nil {
			return nil, errLoadFolders
		}
		inMemoryRepository = NewInMemoryRepository(&folders)
	}

	updatedFolder, errUpdateFolder := inMemoryRepository.Update(ctx, folder)
	if errUpdateFolder != nil {
		return nil, errUpdateFolder
	}
	folderPtrs, errGetFolders := inMemoryRepository.Get(ctx, ListFolderParams{ListParams: generics.ListParams{Deleted: model.YesOrNo}})
	if errGetFolders != nil {
		return nil, errGetFolders
	}
	var folders []Folder
	for _, folderPtr := range folderPtrs {
		folders = append(folders, *folderPtr)
	}
	return updatedFolder, m.encode(&folders)
}

func (m FileRepository) Create(ctx context.Context, folder Folder) (*Folder, error) {
	if m.inMemoryRepository != nil {
		createdFolder, errCreateFolder := m.inMemoryRepository.Create(ctx, folder)
		if errCreateFolder != nil {
			return nil, errors.Wrapf(errCreateFolder, "Error creating folder with ID of %d in memory", folder.ID)
		}

		folderPtrs, errGetFolders := m.inMemoryRepository.Get(ctx, ListFolderParams{ListParams: generics.ListParams{Deleted: model.YesOrNo}})
		if errGetFolders != nil {
			return nil, errGetFolders
		}
		var folders []Folder
		for _, folderPtr := range folderPtrs {
			folders = append(folders, *folderPtr)
		}
		errencode := m.encode(&folders)
		return createdFolder, errencode
	}

	// Done this way because file may have a duplicate entry and needs to be
	// loaded to check
	var folders []Folder
	errDecode := m.decode(&folders)
	if errDecode != nil {
		return nil, errDecode
	}

	inMemoryRepository := NewInMemoryRepository(&folders)
	createdFolder, errCreateFolder := inMemoryRepository.Create(ctx, folder)
	if errCreateFolder != nil {
		return nil, errors.Wrapf(errCreateFolder, "Error creating folder with ID of %d in memory", folder.ID)
	}

	errEncode := m.encode(&folders)
	if errEncode != nil {
		return nil, errEncode
	}

	return createdFolder, nil
}

func (m FileRepository) Count(ctx context.Context, params ListFolderParams) (uint, error) {
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
		folders, errLoadFolders := m.Load(ctx)
		if errLoadFolders != nil {
			return errLoadFolders
		}
		inMemoryRepository = NewInMemoryRepository(&folders)
	}

	errDelete := inMemoryRepository.Delete(ctx, id, forceDelete)
	if errDelete != nil {
		return errDelete
	}

	folderPtrs, errGetFolders := inMemoryRepository.Get(ctx, ListFolderParams{ListParams: generics.ListParams{Deleted: model.YesOrNo}})
	if errGetFolders != nil {
		return errGetFolders
	}
	var folders []Folder
	for _, folderPtr := range folderPtrs {
		folders = append(folders, *folderPtr)
	}
	return m.encode(&folders)
}

func (m FileRepository) encode(data *[]Folder) error {
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

func (m FileRepository) decode(data *[]Folder) error {
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
