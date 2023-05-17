/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"database/sql"

	"github.com/gofrs/uuid"

	// driver package for postgresql just needs import
	"github.com/lib/pq"

	"github.com/alexflint/go-arg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/utils"
)

// SQLConfig the sql connection args for our postgresql db connection
type SQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type AuthentikBlueprintInstance struct {
	Created         time.Time       `json:"created" yaml:"created"`
	LastUpdated     time.Time       `json:"last_updated" yaml:"last_updated"`
	Managed         string          `json:"managed" yaml:"managed"`
	InstanceUUID    uuid.UUID       `json:"instance_uuid" yaml:"instance_uuid"`
	Name            string          `json:"name" yaml:"name"`
	Metadata        json.RawMessage `json:"metadata" yaml:"metadata"`
	Path            string          `json:"path" yaml:"path"`
	Context         json.RawMessage `json:"context" yaml:"context"`
	LastApplied     time.Time       `json:"last_applied" yaml:"last_applied"`
	LastAppliedHash string          `json:"last_applied_hash" yaml:"last_applied_hash"`
	Status          string          `json:"status" yaml:"status"`
	Enabled         bool            `json:"enabled" yaml:"enabled"`
	ManagedModels   []string        `json:"managed_models" yaml:"managed_models"`
	Content         string          `json:"content" yaml:"content"`
}

// NewSQLConfig best effort to generate a connection config based on env variables and system
func NewSQLConfig() *SQLConfig {
	// TODO populate with real values from go-arg
	return &SQLConfig{
		Host:     "postgres",
		Port:     5432,
		User:     "postgres",
		Password: "MIwHsckSqhCli0KCEmq5RZDld744vP", // this is the password from example secret in docs docs
		DBName:   "authentik",
		SSLMode:  "disable",
	}
}

// SQLConnect gets and test a basic SQL connection to our postgres database specifically
func SQLConnect(config *SQLConfig) (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, strconv.Itoa(config.Port), config.User, config.Password, config.DBName, config.SSLMode)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// AkBlueprintReconciler reconciles a AkBlueprint object
type AkBlueprintReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=akblueprints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=akblueprints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=akm.goauthentik.io,resources=akblueprints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AkBlueprintReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := klog.FromContext(ctx)

	// PARSE OPTIONS
	// TODO: pass them in rather than read continuously
	o := utils.Opts{}
	arg.MustParse(&o)
	l.Info(utils.PrettyPrint(o))

	// GET CRD
	crd := &akmv1a1.AkBlueprint{}
	err := r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("AkBlueprint disappeared.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "AkBlueprint irretrievable, Retrying.")
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found AkBlueprint `%v` in `%v`.", crd.Name, crd.Namespace))

	// CREATE CONFIGMAP
	name := fmt.Sprintf("bp-%v-%v", crd.Namespace, crd.Name)
	cmWant, err := r.configForBlueprint(crd, name, crd.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	cm := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: name, Namespace: crd.Namespace}, cm)
	if err != nil && errors.IsNotFound(err) {
		// configmap was not found rety and notify the user
		l.Info(fmt.Sprintf("Not found. Creating configmap `%v` in `%v`", name, crd.Namespace))
		err = r.Create(ctx, cmWant)
		if err != nil {
			l.Error(err, fmt.Sprintf("Failed to create configmap `%v` in `%v`", name, crd.Namespace))
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		// something went wrong with fetching the config map could be fatal
		l.Error(err, fmt.Sprintf("Failed to get configmap `%v` in `%v`", name, crd.Namespace))
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found configmap %v in %v", name, crd.Namespace))
	//check configmap matches what we want it to be by updating it
	r.Update(ctx, cmWant)
	if err != nil {
		// something went wrong with updating the deployment
		l.Error(err, fmt.Sprintf("Failed to update configmap %v in %v", name, crd.Namespace))
		return ctrl.Result{}, err
	}

	// POPULATE DATABASE ROW STRUCT
	id, err := uuid.NewV7()
	if err != nil {
		return ctrl.Result{}, err
	}
	crdyml, err := yaml.Marshal(crd.Spec.Blueprint)
	if err != nil {
		return ctrl.Result{}, err
	}
	metajson, err := json.Marshal(&crd.Spec.Blueprint.Metadata)
	if err != nil {
		return ctrl.Result{}, err
	}
	metamsg := json.RawMessage(metajson)
	row := AuthentikBlueprintInstance{
		Created:     time.Now(),
		LastUpdated: time.Now(),
		//Managed:      "",
		InstanceUUID: id,
		Name:         crd.Spec.Blueprint.Metadata.Name,
		Metadata:     metamsg,
		//Path:         "SomePath",
		Context:     json.RawMessage(`{}`),
		LastApplied: time.Now(),
		//LastAppliedHash: "text",
		Status:        "unknown",
		Enabled:       true,
		ManagedModels: []string{},
		Content:       string(crdyml),
	}

	// SETUP DB CONNECTION
	cfg := NewSQLConfig()
	l.Info(fmt.Sprintf("Connecting to postgresql at %v in %v...", cfg.Host, crd.Namespace))
	db, err := SQLConnect(cfg)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer db.Close()
	l.Info(fmt.Sprintf("Connected to postgresql at %v in %v", cfg.Host, crd.Namespace))

	// QUERY DB
	tableName := "authentik_blueprints_blueprintinstance"
	columnName := "path"
	current, err := queryRowByColumnValue(db, tableName, columnName, crd.Spec.File)
	if err != nil {
		return ctrl.Result{}, err
	}
	if current != nil {
		l.Info(fmt.Sprintf("In postgresql at `%v` in ns `%v` found `%v`", cfg.Host, crd.Namespace, current))
		// found so update / check it is what we want
		err = updateRowBySchema(db, &row, tableName)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		l.Info(fmt.Sprintf("Adding blueprint to postgresql at `%v` in ns `%v`", cfg.Host, crd.Namespace))
		// missing so add
		err = addRowBySchema(db, &row, tableName)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func addRowBySchema(db *sql.DB, row *AuthentikBlueprintInstance, tableName string) error {
	stmt := fmt.Sprintf("INSERT INTO %v (created,last_updated,managed,instance_uuid,name,metadata,path,context,last_applied,last_applied_hash,status,enabled,managed_models,content) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", tableName)

	var managed interface{}
	if row.Managed != "" {
		managed = row.Managed
	} else {
		managed = nil // Insert NULL for the field
	}

	_, err := db.Exec(stmt,
		&row.Created,
		&row.LastUpdated,
		managed,
		&row.InstanceUUID,
		&row.Name,
		&row.Metadata,
		&row.Path,
		&row.Context,
		&row.LastApplied,
		&row.LastAppliedHash,
		&row.Status,
		&row.Enabled,
		pq.Array(row.ManagedModels),
		&row.Content)
	return err
}

func updateRowBySchema(db *sql.DB, row *AuthentikBlueprintInstance, tableName string) error {
	return nil
}

func queryRowByColumnValue(db *sql.DB, tableName string, columnName string, columnValue string) (*AuthentikBlueprintInstance, error) {
	// TODO: use db.Query args rather than fmt
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", tableName, columnName)

	row := db.QueryRow(query, columnValue)

	// Create a new struct instance to hold the row data
	var result AuthentikBlueprintInstance

	// Scan the row data into the struct fields
	err := row.Scan(
		&result.Created, &result.LastUpdated, &result.Managed, &result.InstanceUUID, &result.Name, &result.Metadata, &result.Path, &result.Context, &result.LastApplied, &result.LastAppliedHash, &result.Status, &result.Enabled, &result.ManagedModels, &result.Content,
	)
	//err := row.Scan(&rowData.ID, &rowData.Name, &rowData.Email, &rowData.JSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows found with the specified column value
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

// configForBlueprint generates a configmap spec from a given blueprint that contains the blueprint data as a kube-native configmap to mount into our deployment later.
func (r *AkBlueprintReconciler) configForBlueprint(crd *akmv1a1.AkBlueprint, name string, namespace string) (*corev1.ConfigMap, error) {
	// create the map of key values for the data in configmap from blueprint contents
	cleanFP := filepath.Clean(crd.Spec.File)
	var dataMap = make(map[string]string)
	// set the key to be the filename and extension from path
	// set data to be the blueprint string
	b, err := json.Marshal(crd.Spec.Blueprint)
	if err != nil {
		return nil, err
	}
	dataMap[filepath.Base(cleanFP)] = string(b)

	// create annotation for destination path
	var annMap = make(map[string]string)
	annMap["akm.goauthentik/v1alpha1"] = filepath.Dir(cleanFP)

	cm := corev1.ConfigMap{
		// Metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annMap,
		},
		Data: dataMap,
	}
	// set that we are controlling this resource
	ctrl.SetControllerReference(crd, &cm, r.Scheme)
	return &cm, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AkBlueprintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akmv1a1.AkBlueprint{}).
		Complete(r)
}
