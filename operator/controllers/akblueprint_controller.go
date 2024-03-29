/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"database/sql"

	"github.com/gofrs/uuid"

	// driver package for postgresql just needs import
	"github.com/lib/pq"

	"github.com/alexflint/go-arg"
	yaml_v3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
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

	//// GET CRD WORKAROUND / MONKEY PATCH
	//// currently the controller-runtime does not support yaml.v3 unmarshalling of custom YAML tags
	//// When you call r.Get the controller-runtime will try to unmarshal the yaml, we need to NOT do this
	//// since it is based on a yaml.v2 fork. Instead we use yaml.v3 by calling kube directly ourselved to get
	//// the yaml bytes we are interested in.
	//clientset, err := k8s.NewClient(req.Namespace)
	//if err != nil {
	//	l.Error(err, "Failed to create k8s client. Retrying.")
	//	return ctrl.Result{}, err
	//}
	//fmt.Printf("clientset: %v\n", clientset)

	//// https://docs.openshift.com/container-platform/3.11/admin_guide/custom_resource_definitions.html
	//// API endpoints are defined by the following template: /apis/<spec:group>/<spec:version>/<scope>/*/<names-plural>/...
	////crdBytes, err := clientset.RESTClient().Get().RequestURI("/openapi/v3").DoRaw(ctx)
	//apiEndpoint := "/apis/akm.goauthentik.io/v1alpha1/namespaces/" + req.Namespace + "/akblueprints/" + req.Name
	//fmt.Printf("apiEndpoint: %v\n", apiEndpoint)
	//crdBytes, err := clientset.RESTClient().Get().RequestURI(apiEndpoint).DoRaw(ctx)
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		// Request object not found, could have been deleted after reconcile request.
	//		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
	//		// Return and don't requeue
	//		l.Info("AkBlueprint trigger deletion...")
	//		// I prefer early return but we actually cant do that here as we need the CRD and the DB to be ready
	//		// the DB requires the CRD but I also dont want to open the DB twice.
	//		markedForDeletion = true
	//	} else {
	//		// Error reading the object - requeue the request.
	//		l.Error(err, "AkBlueprint trigger irretrievable, Retrying.")
	//		return ctrl.Result{}, err
	//	}
	//} else {

	//	l.Info("AkBlueprint trigger.")
	//}
	//fmt.Printf("crdBytes: %v\n", string(crdBytes))

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

	// FIND RELEVANT Ak resources
	// we know only one Ak resource should be present
	// so we can search for it in our namespace!
	// TODO: convert to watched namespace when full pipeline is ready
	list, err := r.ListAk(o.OperatorNamespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(list) > 1 {
		err = errors.NewConflict(
			schema.GroupResource{
				Group:    "akm.goauthentik.io",
				Resource: crd.Name,
			},
			fmt.Sprintf("Too many Ak resources, cant decide between them `%v`.", list),
			fmt.Errorf("Too many relevant Ak resources"))
		return ctrl.Result{}, err
	} else if len(list) == 0 {
		err = errors.NewNotFound(
			schema.GroupResource{
				Group:    "akm.goauthentik.io",
				Resource: crd.Name,
			},
			fmt.Sprintf("No relevant Ak resource found."))
		return ctrl.Result{}, err
	}
	ak := list[0]
	l.Info(fmt.Sprintf("Found relevant Ak resource."))

	// FIND AK RESOURCES RELEASED VALUES
	// We find the Ak resources values so we can ensure we are searching for the correct
	// secret
	values, err := r.GetReleasedValues(o.WatchedNamespace, ak.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	section, ok := values["secret"].(map[string]interface{})
	if !ok {
		// TODO: Populate error
		return ctrl.Result{}, err
	}
	secretName, ok := section["name"].(string)
	if !ok {
		// TODO: Populate error
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found release secret name `%v`", secretName))

	// SCRAPE AUTHENTIK RELEASED SECRET
	secret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: req.NamespacedName.Namespace}, secret)
	if err != nil {
		return ctrl.Result{}, err
	}
	l.Info(fmt.Sprintf("Found release secret `%v`", secretName))

	// SETUP DB CONNECTION
	cfg := &utils.SQLConfig{
		Host:     "postgres",
		Port:     5432,
		User:     "postgres",
		Password: string(secret.Data["postgresPassword"][:]),
		DBName:   "authentik",
		SSLMode:  "disable",
	}
	//cfg := r.NewSQLConfig()
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

	// DECODE CRD INTO STRUCTURED BLUEPRINT
	bp := &akmv1a1.BP{}
	decoder := yaml_v3.NewDecoder(bytes.NewReader([]byte(crd.Spec.Blueprint)))
	if err := decoder.Decode(bp); err != nil {
		return ctrl.Result{}, err
	}

	// CREATE CONFIGMAP
	if crd.Spec.StorageType == "file" {
		name := fmt.Sprintf("bp-%v-%v", crd.Namespace, crd.Name)
		cmWant, err := r.configForBlueprint(crd, name, crd.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		l.Info(fmt.Sprintf("Searching for configmap `%v` in `%v`...", name, crd.Namespace))
		cm := &corev1.ConfigMap{}
		err = r.Get(ctx, types.NamespacedName{Name: name, Namespace: crd.Namespace}, cm)
		if err != nil && errors.IsNotFound(err) {
			// configmap was not found create and notify the user
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

	metajson, err := json.Marshal(&bp.Metadata)
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

	if crd.Spec.StorageType == "internal" {
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
	// apply regex substitution to remove quotes ['"](?P<content>\!.*)['"] -> ${content}
	// this is required since authentiks python yaml parser doesn't like quotes on
	// their custom yaml tags so we have to ensure they are stripped here for consistency
	regexPatterns := map[string]string{
		`['"](?P<content>\!.*)['"]`: "${content}", // This strips the quotes from special yaml tags
		//`['"](?P<content>null)['"]`: "${content}", // This strips the quotes from "null"
		//`['"](?P<content>true)['"]`:  "${content}", // This strips the quotes from "true"
		//`['"](?P<content>false)['"]`: "${content}", // This strips the quotes from "false"
	}
	cleanedBlueprint := regexSubstituteMap(regexPatterns, string(crd.Spec.Blueprint))
	// set the configmap key to be the file name we want it to be mounted as for the volume mounts
	dataMap[filepath.Base(cleanFP)] = cleanedBlueprint

	var annMap = make(map[string]string)
	annMap["akm.goauthentik.io/path"] = filepath.Dir(cleanFP)

	// create label to specifically identify blueprint related configmaps
	var labelMap = make(map[string]string)
	labelMap["akm.goauthentik.io/type"] = "blueprint"
	labelMap["akm.goauthentik.io/blueprint"] = crd.Name

	cm := corev1.ConfigMap{
		// Metadata
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labelMap,
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

// regexSubstituteMap takes in a map[string]string of regex patterns as keys and regex replacements as values
// this is then applied to a given string by iterating over the keys to find matches and replacing the values
func regexSubstituteMap(patterns map[string]string, data string) string {
	result := data
	for pattern, replacement := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, replacement)
	}
	return result
}
