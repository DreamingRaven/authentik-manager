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
	"strings"
	"time"

	"database/sql"

	"github.com/gofrs/uuid"

	// driver package for postgresql just needs import
	"github.com/lib/pq"

	"github.com/alexflint/go-arg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	akmv1a1 "gitlab.com/GeorgeRaven/authentik-manager/operator/api/v1alpha1"
	"gitlab.com/GeorgeRaven/authentik-manager/operator/utils"
)

type AuthentikBlueprintInstance struct {
	Created         time.Time       `json:"created"`
	LastUpdated     time.Time       `json:"last_updated"`
	Managed         string          `json:"managed"`
	InstanceUUID    uuid.UUID       `json:"instance_uuid"`
	Name            string          `json:"name"`
	Metadata        json.RawMessage `json:"metadata"`
	Path            string          `json:"path"`
	Context         json.RawMessage `json:"context"`
	LastApplied     time.Time       `json:"last_applied"`
	LastAppliedHash string          `json:"last_applied_hash"`
	Status          string          `json:"status"`
	Enabled         bool            `json:"enabled"`
	ManagedModels   []string        `json:"managed_models"`
	Content         string          `json:"content"`
}

// ListAk returns a list of Ak resources in the given namespace
func (r *AkBlueprintReconciler) ListAk(namespace string) ([]*akmv1a1.Ak, error) {
	list := &akmv1a1.AkList{}
	opts := &client.ListOptions{
		Namespace: namespace,
	}
	err := r.List(context.TODO(), list, opts)
	if err != nil {
		return nil, err
	}
	// Unpack into an actual list
	resources := make([]*akmv1a1.Ak, len(list.Items))
	for i, item := range list.Items {
		resources[i] = &item
	}

	return resources, nil
}

// AkBlueprintReconciler reconciles a AkBlueprint object
type AkBlueprintReconciler struct {
	utils.ControlBase
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
	//l.Info(utils.PrettyPrint(o))
	tableName := "authentik_blueprints_blueprintinstance"
	markedForDeletion := false

	// GET CRD
	crd := &akmv1a1.AkBlueprint{}
	err := r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			l.Info("AkBlueprint trigger deletion...")
			// I prefer early return but we actually cant do that here as we need the CRD and the DB to be ready
			// the DB requires the CRD but I also dont want to open the DB twice.
			markedForDeletion = true
		} else {
			// Error reading the object - requeue the request.
			l.Error(err, "AkBlueprint trigger irretrievable, Retrying.")
			return ctrl.Result{}, err
		}
	} else {

		l.Info("AkBlueprint trigger.")
	}

	// SETUP DB CONNECTION
	cfg := r.NewSQLConfig()
	l.Info(fmt.Sprintf("Connecting to postgresql at %v in %v...", cfg.Host, req.NamespacedName.Namespace))
	db, err := utils.SQLConnect(cfg)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer db.Close()
	l.Info("Connected")

	// DELETING DB ROW FOR REMOVED CRD
	if markedForDeletion {
		deleteColumnValues := map[string]interface{}{
			//"path": "",
			"name": req.NamespacedName.Name,
		}
		l.Info("Deleting...")
		result, err := deleteRowsByColumnValues(db, tableName, deleteColumnValues)
		if err != nil {
			return ctrl.Result{}, err
		}
		count, err := (*result).RowsAffected()
		if count == 0 {
			// no rows deleted
			l.Info("Nothing deleted")
		} else if count == 1 {
			l.Info("Deleted")
		} else {
			l.Info("Multiple deleted")
		}
		return ctrl.Result{}, nil
	}

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
	rowDesire := AuthentikBlueprintInstance{
		Created:     time.Now(),
		LastUpdated: time.Now(),
		//Managed:      "",
		InstanceUUID: id,
		Name:         crd.Name,
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

	// QUERY DB
	searchColumnValues := map[string]interface{}{
		// the purpose of searching with paths is to deal with default blueprint overrides
		// we need some way to enable people to overwrite by path as well as by name
		// since this is what would happen when we overwrite files the name can change
		// if we end up getting more than 1 then we throw an error
		// blueprints added internally are not stored with paths so it should have no effect
		// after the merge
		"path": crd.Spec.File,
		"name": crd.Name,
	}
	rows, err := searchRowsByColumnValues(db, tableName, searchColumnValues)
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(rows) == 0 {
		// IF NOT FOUND CREATE
		l.Info(fmt.Sprintf("No db blueprint found creating `%v`", crd.Name))
		err = addRowBySchema(db, &rowDesire, tableName)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if len(rows) == 1 {
		// IF 1 FOUND RECONCILE IT
		l.Info(fmt.Sprintf("Db blueprint found reconciling `%v`", rows[0].Name))
		result, err := updateRowByColumns(db, tableName, searchColumnValues, rowDesire)
		if err != nil {
			return ctrl.Result{}, err
		}
		count, err := (*result).RowsAffected()
		if count == 0 {
			l.Info("Nothing modified")
		} else if count == 1 {
			l.Info("Modified")
		} else {
			l.Info("Multiple modified")
		}
	} else {
		// IF MULTIPLE FOUND THROW
		err = errors.NewConflict(
			schema.GroupResource{
				Group:    "akm.goauthentik.io",
				Resource: "AkBlueprint",
			},
			fmt.Sprintf("Too many (%v) db blueprints found cannot reconcile `%v`", len(rows), rows),
			nil,
		)
		return ctrl.Result{}, err
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

func updateRowByColumns(db *sql.DB, tableName string, columnValues map[string]interface{}, newValues AuthentikBlueprintInstance) (*sql.Result, error) {
	var conditions []string
	var args []interface{}

	var setValues []string
	index := 1

	// Construct the WHERE clause based on the column values
	for column, value := range columnValues {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, index))
		args = append(args, value)
		index++
	}

	// Construct the SET clause with the new values
	index = len(columnValues) + 1
	for field, value := range map[string]interface{}{
		"created":           nil, // Keep the original value
		"last_updated":      newValues.LastUpdated,
		"managed":           nil,
		"instance_uuid":     nil,
		"name":              newValues.Name,
		"metadata":          newValues.Metadata,
		"path":              newValues.Path,
		"context":           newValues.Context,
		"last_applied":      newValues.LastApplied,
		"last_applied_hash": newValues.LastAppliedHash,
		"status":            nil,
		"enabled":           newValues.Enabled,
		"managed_models":    pq.Array(newValues.ManagedModels),
		"content":           newValues.Content,
	} {
		if value != nil {
			setValues = append(setValues, fmt.Sprintf("%s = $%d", field, index))
			args = append(args, value)
			index++
		}
	}

	// Construct the UPDATE query
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, strings.Join(setValues, ", "), strings.Join(conditions, " AND "))

	// Execute the query
	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func updateRowBySchema(db *sql.DB, row *AuthentikBlueprintInstance, tableName string) error {
	return nil
}

func deleteRowsByColumnValues(db *sql.DB, tableName string, columnValues map[string]interface{}) (*sql.Result, error) {
	// Build the WHERE clause using the column names and values
	var conditions []string
	var args []interface{}

	index := 1
	for column, value := range columnValues {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, index))
		args = append(args, value)
		index++
	}

	// Construct the DELETE statement
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, strings.Join(conditions, " AND "))

	// Execute the DELETE statement
	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	//rowsAffected, err := result.RowsAffected()
	//if err != nil {
	//	return 0, err
	//}

	return &result, nil
}

func searchRowsByColumnValues(db *sql.DB, tableName string, columnValues map[string]interface{}) ([]AuthentikBlueprintInstance, error) {
	// Construct the WHERE clause based on the column values
	var whereClauses []string
	var args []interface{}
	i := 1
	for column, value := range columnValues {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", column, i))
		args = append(args, value)
		i++
	}
	whereClause := strings.Join(whereClauses, " OR ")

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, whereClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AuthentikBlueprintInstance

	for rows.Next() {
		var managed sql.NullString
		var result AuthentikBlueprintInstance
		err := rows.Scan(
			&result.Created,
			&result.LastUpdated,
			&managed,
			&result.InstanceUUID,
			&result.Name,
			&result.Metadata,
			&result.Path,
			&result.Context,
			&result.LastApplied,
			&result.LastAppliedHash,
			&result.Status,
			&result.Enabled,
			pq.Array(&result.ManagedModels),
			&result.Content,
		)
		if err != nil {
			return nil, err
		}
		if managed.Valid {
			result.Managed = managed.String
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
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
