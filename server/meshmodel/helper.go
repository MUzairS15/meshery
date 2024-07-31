package meshmodel

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/layer5io/meshery/server/helpers"
	"github.com/layer5io/meshery/server/helpers/utils"
	"github.com/layer5io/meshery/server/models"
	"github.com/layer5io/meshery/server/models/meshmodel/core"
	"github.com/layer5io/meshkit/logger"
	"github.com/layer5io/meshkit/models/meshmodel/core/v1alpha2"
	_models "github.com/layer5io/meshkit/models/meshmodel/core/v1beta1"

	"github.com/meshery/schemas/models/v1beta1/connection"
	"github.com/meshery/schemas/models/v1beta1/component"

	meshmodel "github.com/layer5io/meshkit/models/meshmodel/registry"
	mutils "github.com/layer5io/meshkit/utils"
	"github.com/pkg/errors"
)

var ArtifactHubComponentsHandler = _models.ArtifactHub{} //The components generated in output directory will be handled by kubernetes
var ModelsPath = "../meshmodel"
var RelativeRelationshipsPath = "relationships"

type EntityRegistrationHelper struct {
	handlerConfig    *models.HandlerConfig
	regManager       *meshmodel.RegistryManager
	componentChan    chan component.ComponentDefinition
	relationshipChan chan v1alpha2.RelationshipDefinition
	errorChan        chan error
	log              logger.Handler
}

func NewEntityRegistrationHelper(hc *models.HandlerConfig, rm *meshmodel.RegistryManager, log logger.Handler) *EntityRegistrationHelper {
	return &EntityRegistrationHelper{
		handlerConfig:    hc,
		regManager:       rm,
		componentChan:    make(chan component.ComponentDefinition),
		relationshipChan: make(chan v1alpha2.RelationshipDefinition),
		errorChan:        make(chan error),
		log:              log,
	}
}

// seed the local meshmodel components
func (erh *EntityRegistrationHelper) SeedComponents() {
	// Watch channels and register components and relationships with the registry manager
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go erh.watchComponents(ctx)

	models, err := os.ReadDir(ModelsPath)
	if err != nil {
		erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while reading directory for generating components"), ModelsPath)
		return
	}

	relationships := make([]string, 0)

	// change to queue approach to register comps, relationships and policies
	// Read component and relationship definitions from files and send them to respective channels

	for _, model := range models {
		// fmt.Println("model: ", model.Name())
		if model.Name() != "kubernetes" {
			continue
		}
		modelVersionsDirPath := filepath.Join(ModelsPath, model.Name())
		// contains all versions for the model
		modelVersionsDir, err := os.ReadDir(modelVersionsDirPath)
		if err != nil {
			erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while reading directory for registering components"), modelVersionsDirPath)
			continue
		}

		fmt.Println("REACHED 83 ")

		for _, version := range modelVersionsDir {
			modelDefVersionsDirPath := filepath.Join(modelVersionsDirPath, version.Name())

			// contains all def versions for a particular version of a model.
			modelDefVersionsDir, err := os.ReadDir(modelDefVersionsDirPath)
			if err != nil {
				erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while reading directory for registering components"), modelVersionsDirPath)
			}
			fmt.Println("REACHED 93 ")
			for _, defVersion := range modelDefVersionsDir {
				defPath := filepath.Join(modelDefVersionsDirPath, defVersion.Name())

				entities, err := os.ReadDir(defPath)
				if err != nil {
					erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while reading directory for registering components"), modelVersionsDirPath)
					continue
				}
				fmt.Println("REACHED 102")
				for _, entity := range entities {
					entityPath := filepath.Join(defPath, entity.Name())
					if entity.IsDir() {
						switch entity.Name() {
						case "relationships":
							relationships = append(relationships, entityPath)
						case "policies":
						default:
							fmt.Println("REACHED 111 ")
							erh.generateComponents(entityPath) // register components first
						}
					}
				}
			}
		}
	}
	for _, relationship := range relationships {
		erh.generateRelationships(relationship)
	}
}

// reads component definitions from files and sends them to the component channel
func (erh *EntityRegistrationHelper) generateComponents(pathToComponents string) {
	path, err := filepath.Abs(pathToComponents)
	fmt.Println("REACHED 127 ", pathToComponents)
	if err != nil {
		erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while getting absolute path for generating components"), pathToComponents)
		return
	}

	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		if !info.IsDir() {
			// Read the component definition from file
			var comp component.ComponentDefinition
			byt, err := os.ReadFile(path)
			fmt.Println("REACHED 142 ")
			if err != nil {
				erh.errorChan <- mutils.ErrReadFile(errors.Wrapf(err, fmt.Sprintf("unable to read file at %s", path)), path)
				return nil
			}

			err = json.Unmarshal(byt, &comp)
			fmt.Println("REACHED 149 ", comp.DisplayName, err)
			if err != nil {
				erh.errorChan <- mutils.ErrMarshal(errors.Wrapf(err, fmt.Sprintf("unmarshal json failed for %s", path)))
				return nil
			}

			// Only register components that have been marked as published

			// && comp.Status == "enabled" add this condition

			// Generate SVGs for the component and save them on the file system
			utils.WriteSVGsOnFileSystem(&comp)
			erh.componentChan <- comp

		}
		return nil
	})
	if err != nil {
		erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while generating components"), pathToComponents)
	}
}

// reads relationship definitions from files and sends them to the relationship channel
func (erh *EntityRegistrationHelper) generateRelationships(pathToComponents string) {
	path, err := filepath.Abs(pathToComponents)
	if err != nil {
		erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while getting absolute path for generating relationships"), pathToComponents)
		return
	}

	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if !info.IsDir() {
			var rel v1alpha2.RelationshipDefinition
			byt, err := os.ReadFile(path)
			if err != nil {
				erh.errorChan <- mutils.ErrReadFile(errors.Wrapf(err, fmt.Sprintf("unable to read file at %s", path)), path)
				return nil
			}
			err = json.Unmarshal(byt, &rel)
			if err != nil {
				erh.errorChan <- mutils.ErrUnmarshal(errors.Wrapf(err, fmt.Sprintf("unmarshal json failed for %s", path)))
				return nil
			}
			erh.relationshipChan <- rel
		}
		return nil
	})
	if err != nil {
		erh.errorChan <- mutils.ErrReadDir(errors.Wrapf(err, "error while generating relationships"), pathToComponents)
	}
}

// watches the component and relationship channels for incoming definitions and registers them with the registry manager
// If an error occurs, it logs the error
func (erh *EntityRegistrationHelper) watchComponents(ctx context.Context) {
	var err error
	var isModelError bool
	var isRegistrantError bool
	for {
		select {
		case comp := <-erh.componentChan:
			fmt.Println("comp----------------------: ", comp.DisplayName)
			connectionKind := comp.Model.Registrant.Kind
			isRegistrantError, isModelError, err = erh.regManager.RegisterEntity(connection.Connection{
				Kind: connectionKind,
			}, &comp)
			if err != nil {
				err = core.ErrRegisterEntity(err, string(comp.Type()), comp.DisplayName)
			}
			helpers.HandleError(connection.Connection{
				Kind: connectionKind,
			}, &comp, err, isModelError, isRegistrantError)
		// case rel := <-erh.relationshipChan:
		// 	connectionKind := rel.Model.Connection.Kind
		// 	isRegistrantError, isModelError, err = erh.regManager.RegisterEntity(v1beta1.Host{
		// 		connectionKind: connectionKind,
		// 	}, &rel)
		// 	helpers.HandleError(v1beta1.Host{
		// 		connectionKind: connectionKind,
		// 	}, &rel, err, isModelError, isRegistrantError)
		// 	if err != nil {
		// 		err = core.ErrRegisterEntity(err, string(rel.Type()), rel.Kind)
		// 	}

		//Watching and logging errors from error channel
		case mhErr := <-erh.errorChan:
			if err != nil {
				erh.log.Error(mhErr)
			}

		case <-ctx.Done():
			erh.registryLog()
			return
		}

		if err != nil {
			// go func() { erh.errorChan <- err }()
		}
	}
}
