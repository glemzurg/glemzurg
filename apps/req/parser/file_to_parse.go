package parser

import (
	"path/filepath"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	_EXT_MODEL          = ".model"
	_EXT_GENERALIZATION = ".generalization"
	_EXT_ACTOR          = ".actor"
	_EXT_DOMAIN         = ".domain"
	_EXT_CLASS          = ".class"
	_EXT_USE_CASE       = ".uc"
)

const (
	_PATH_ACTORS = "actors" // The actors path under models (will not be treated as a domain).
)

var _extSortValue = map[string]int{
	_EXT_MODEL:          10, // Higher values sort first.
	_EXT_GENERALIZATION: 9,
	_EXT_ACTOR:          8,
	_EXT_DOMAIN:         7,
	_EXT_CLASS:          5,
	_EXT_USE_CASE:       3,
}

// The data from walking the file tree.
// Should have enough information to parse everything.
type fileToParse struct {
	ModelPath string
	PathRel   string
	PathAbs   string
	// Derive these values.
	FileType       string
	Generalization string // The generalization can be determined from the file extension but can be found anywhere.
	Actor          string // The user can be determined from the path. It is a filename under the model's user folder.
	Domain         string // The domain can be determined from the path. It is the folder just under the model folder.
	Class          string // The class can be determined from the path. It is a filename under a domain folder.
	UseCase        string // The use case can be determined from the path. It is a filename under a domain folder.
}

func newFileToParse(modelPath, pathRel, pathAbs string) (toParse fileToParse, err error) {

	// Get the extention, that is the file type.
	fileType := filepath.Ext(pathRel)

	// If this is a generalization, then the filename is the details of the generalization.
	// Where it is in the file structure is not important.
	generalization := ""
	if fileType == _EXT_GENERALIZATION {
		baseName := filepath.Base(pathRel)
		// Actor must be unique in this model.
		generalization = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	}

	// If this is a user, then the filename is the user.
	// Where it is in the file structure is not important.
	actor := ""
	if fileType == _EXT_ACTOR {
		baseName := filepath.Base(pathRel)
		// Actor must be unique in this model.
		actor = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	}

	// If there is a first part to the path, it is a domain. Unless it is "users/"
	domain := ""
	// Models are above domains.
	if fileType != _EXT_MODEL {
		pathRelParts := strings.Split(pathRel, string(filepath.Separator))
		if len(pathRelParts) > 0 {
			// The actors path is along side domains but is not one.
			if pathRelParts[0] != _PATH_ACTORS {
				domain = pathRelParts[0]
			}
		}
	}

	// If this is a class, then the filename is the class.
	class := ""
	if fileType == _EXT_CLASS {
		baseName := filepath.Base(pathRel)
		// Class must be unique in this model.
		// The same class could be in different domains (illustrating different facets of an entity).
		class = domain + "/" + strings.TrimSuffix(baseName, filepath.Ext(baseName))
	}

	// If this is a use case, then the filename is the use case.
	useCase := ""
	if fileType == _EXT_USE_CASE {
		baseName := filepath.Base(pathRel)
		// Use case must be unique in this model.
		// The same class could be in different domains (illustrating different facets of an entity).
		useCase = domain + "/" + strings.TrimSuffix(baseName, filepath.Ext(baseName))
	}

	toParse = fileToParse{
		ModelPath:      modelPath,
		PathRel:        pathRel,
		PathAbs:        pathAbs,
		FileType:       fileType,
		Generalization: generalization,
		Actor:          actor,
		Domain:         domain,
		Class:          class,
		UseCase:        useCase,
	}

	err = validation.ValidateStruct(&toParse,
		validation.Field(&toParse.ModelPath, validation.Required),
		validation.Field(&toParse.PathRel, validation.Required),
		validation.Field(&toParse.PathAbs, validation.Required),
		validation.Field(&toParse.FileType, validation.Required, validation.In(_EXT_MODEL, _EXT_GENERALIZATION, _EXT_ACTOR, _EXT_DOMAIN, _EXT_CLASS, _EXT_USE_CASE)),
	)
	if err != nil {
		return fileToParse{}, errors.WithStack(err)
	}

	return toParse, nil
}

func (f *fileToParse) String() string {
	return f.FileType + " : " + f.PathRel + " (" + f.PathAbs + ")"
}

func sortFilesToParse(filesToParse []fileToParse) {
	sort.Slice(filesToParse, func(i, j int) bool {
		return lessThanFilesToParse(filesToParse[i], filesToParse[j])
	})
}

func lessThanFilesToParse(fileA, fileB fileToParse) (less bool) {

	// Sort first by file type.
	sortValueA := _extSortValue[fileA.FileType]
	sortValueB := _extSortValue[fileB.FileType]
	if sortValueA != sortValueB {
		return sortValueA > sortValueB // Higher values sort first.
	}

	// Sort next by relative path.
	return fileA.PathRel < fileB.PathRel
}
