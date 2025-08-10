package secrets

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/mholt/archives"
	"github.com/pkg/errors"
	"github.com/risor-io/risor"
	"github.com/risor-io/risor/object"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	secrets2 "hideout/internal/secrets"
	"hideout/services/secrets"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

func (s *Secret) Process(ctx context.Context, secretsSvc *secrets.SecretsService) (string, string, error) {
	if secretsSvc == nil {
		return "", "", errors.New("Secrets service is non-existent")
	}

	secretsList, errGetSecretUIDs := secretsSvc.GetSecrets(ctx, secrets2.ListSecretParams{
		ListParams: generics.ListParams{Deleted: model.No},
	})
	if errGetSecretUIDs != nil {
		return "", "", errGetSecretUIDs
	}

	// Delete current one from the list to avoid self-referencing
	secretsList = slices.DeleteFunc(secretsList, func(dp *secrets2.Secret) bool {
		return dp.UID == s.UID
	})

	var globalValues = map[string]any{}
	// Reference secrets by {{id}} and {{uid}} constructs
	for _, secretEntry := range secretsList {
		globalValues[fmt.Sprintf("{{%s}}", secretEntry.UID)] = secretEntry.Value
		globalValues[fmt.Sprintf("{{%d}}", secretEntry.ID)] = secretEntry.Value
	}

	// Disable dangerous and unnecessary modules
	evaluatedResult, errEvaluate := risor.Eval(ctx, s.Value, risor.WithGlobals(globalValues),
		risor.WithoutGlobals("errors", "exec", "filepath", "http", "net", "os"))
	if errEvaluate != nil {
		return "", "", errEvaluate
	}

	valueType := evaluatedResult.Type()
	switch valueType {
	case object.BOOL:
		return fmt.Sprintf("%t", evaluatedResult.Interface().(bool)), string(object.BOOL), nil
	case object.STRING:
		return fmt.Sprintf("%s", evaluatedResult.Interface().(string)), string(object.STRING), nil
	case object.INT:
		return fmt.Sprintf("%d", evaluatedResult.Interface().(int)), string(object.STRING), nil
	case object.FLOAT:
		return fmt.Sprintf("%f", evaluatedResult.Interface().(float64)), string(object.STRING), nil
	}

	return "", string(valueType), fmt.Errorf("Cannot process result with type of %s", string(valueType))
}

func doubleQuoteEscape(line string) string {
	const doubleQuoteSpecialChars = "\\\n\r\"!$`"
	for _, c := range doubleQuoteSpecialChars {
		toReplace := "\\" + string(c)
		if c == '\n' {
			toReplace = `\n`
		}
		if c == '\r' {
			toReplace = `\r`
		}
		line = strings.Replace(line, string(c), toReplace, -1)
	}
	return line
}

func ExportToDotEnv(ctx context.Context, secrets []Secret) (string, error) {
	if len(secrets) == 0 {
		return "", nil
	}

	lines := make([]string, 0, len(secrets))

	for _, secret := range secrets {
		if d, err := strconv.Atoi(secret.Value); err == nil {
			lines = append(lines, fmt.Sprintf(`%s=%d`, secret.Name, d))
		} else {
			lines = append(lines, fmt.Sprintf(`%s="%s"`, secret.Name, doubleQuoteEscape(secret.Value)))
		}
	}

	return strings.Join(lines, "\n"), nil
}

func ArchiveExport(ctx context.Context, data []byte, archiveType uint, compressionType uint, exportType uint) (string, error) {
	var uuid = gofakeit.UUID()
	exportTypeVal, _ := ExportExtensionsMap[exportType]
	secretsFile := fmt.Sprintf("secrets-%s%s", strings.ReplaceAll(uuid, "-", ""), exportTypeVal)

	secretsTemporaryFile, errCreateTemporaryFile := os.CreateTemp("", secretsFile)
	if errCreateTemporaryFile != nil {
		return "", errCreateTemporaryFile
	}

	defer func() {
		if err := secretsTemporaryFile.Close(); err != nil {
			log.Printf("Error closing secrets temporary file: %v", err)
		}
		if err := os.Remove(secretsTemporaryFile.Name()); err != nil {
			log.Printf("Error removing secrets temporary file: %v", err)
		}
	}()

	_, errWrite := secretsTemporaryFile.Write(data)
	if errWrite != nil {
		return "", errWrite
	}

	// map files on disk to their paths in the archive using default settings (second arg)
	archiveFiles, errCreateArchive := archives.FilesFromDisk(ctx, nil, map[string]string{
		secretsTemporaryFile.Name(): secretsFile,
	})
	if errCreateArchive != nil {
		return "", errCreateArchive
	}

	format := archives.CompressedArchive{}
	switch compressionType {
	case CompressionType_Brotli:
		{
			format.Compression = archives.Brotli{}
			break
		}
	case CompressionType_Bzip2:
		{
			format.Compression = archives.Bz2{}
			break
		}
	case CompressionType_Flate:
		{
			// @TODO Figure out how to do ZIP compression
			return "", apperror.ErrNotImplemented
		}
	case CompressionType_Gzip:
		{
			format.Compression = archives.Gz{}
			break
		}
	case CompressionType_Lz4:
		{
			format.Compression = archives.Lz4{}
			break
		}
	case CompressionType_Lzip:
		{
			format.Compression = archives.Lzip{}
			break
		}
	case CompressionType_Minlz:
		{
			format.Compression = archives.MinLZ{}
			break
		}
	case CompressionType_Snappy:
		{
			format.Compression = archives.Sz{}
			break
		}
	case CompressionType_XZ:
		{
			format.Compression = archives.Xz{}
			break
		}
	case CompressionType_Zlib:
		{
			format.Compression = archives.Zlib{}
			break
		}
	case CompressionType_Zstandard:
		{
			format.Compression = archives.Zstd{}
			break
		}
	}

	switch archiveType {
	case ArchiveType_Tar:
		{
			format.Archival = archives.Tar{}
			break
		}
	case ArchiveType_Zip:
		{
			format.Archival = archives.Zip{}
			break
		}
	}

	archiveTypeVal, _ := ArchiveTypesMap[archiveType]
	secretsArchiveFile := fmt.Sprintf("secrets-%s.%s", strings.ReplaceAll(uuid, "-", ""), archiveTypeVal)

	secretsArchiveTemporaryFile, errCreateSecretsArchiveTemporaryFile := os.CreateTemp("", secretsArchiveFile)
	if errCreateSecretsArchiveTemporaryFile != nil {
		return "", errCreateSecretsArchiveTemporaryFile
	}

	defer func() {
		if err := secretsArchiveTemporaryFile.Close(); err != nil {
			log.Printf("Error closing secrets archive temporary file: %v", err)
		}
	}()

	errArchive := format.Archive(ctx, secretsArchiveTemporaryFile, archiveFiles)
	if errArchive != nil {
		return "", errArchive
	}

	return secretsArchiveTemporaryFile.Name(), nil
}
