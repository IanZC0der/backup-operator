/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorkubecentercomv1beta1 "github.com/IanZC0der/backup-operator/api/v1beta1"
)

// BackUpReconciler reconciles a BackUp object
type BackUpReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	BackupQueue map[string]operatorkubecentercomv1beta1.BackUp
	Wg          sync.WaitGroup
	Tickers     []*time.Ticker
	lock        sync.Mutex
}

// +kubebuilder:rbac:groups=operator.kubecenter.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kubecenter.com,resources=backups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.kubecenter.com,resources=backups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BackUp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *BackUpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var backupK8S operatorkubecentercomv1beta1.BackUp
	err := r.Client.Get(ctx, req.NamespacedName, &backupK8S)
	if err != nil {
		// resource doesn't exist
		if errors.IsNotFound(err) {
			// recource not found, delete from queue
			r.DeleteQueue(req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// resource exists

	// if the spec is the same as the last backup, return
	if lastBackup, ok := r.BackupQueue[backupK8S.Name]; ok {
		if isSame := reflect.DeepEqual(backupK8S.Spec, lastBackup.Spec); isSame {
			return ctrl.Result{}, nil
		}
	}

	//add the backup to the queue
	r.AddQueue(backupK8S)
	return ctrl.Result{}, nil
}

func (r *BackUpReconciler) AddQueue(backup operatorkubecentercomv1beta1.BackUp) {
	if r.BackupQueue == nil {
		r.BackupQueue = make(map[string]operatorkubecentercomv1beta1.BackUp)
	}
	r.BackupQueue[backup.Name] = backup
	// map if not thread-safe, stop the task first
	r.StopTask()
	go r.StartTask()
}

func (r *BackUpReconciler) DeleteQueue(name string) {
	delete(r.BackupQueue, name)

	r.StopTask()
	// map if not thread-safe, stop the task first
	go r.StartTask()
}

// Stop all the tickers
func (r *BackUpReconciler) StopTask() {
	for _, t := range r.Tickers {
		if t != nil {
			t.Stop()
		}
	}

}

/*
 */
func (r *BackUpReconciler) StartTask() {
	for _, backup := range r.BackupQueue {
		if !backup.Spec.Enable {
			backup.Status.Active = false
			r.UpdateStatus(backup)
			continue
		}

		// the expected start time is specified in the spec. compare with the current time and get its actual start time
		actualStartTime := r.GetActualStartTime(backup.Spec.StartTime)
		// get next start time and set the status to active
		backup.Status.Active = true
		nextStartTime := r.GetNextTime(actualStartTime.Seconds())
		backup.Status.NextTime = nextStartTime.Unix()
		r.UpdateStatus(backup)
		ticker := time.NewTicker(actualStartTime)
		r.Tickers = append(r.Tickers, ticker)
		// start a go routine to execute the backup task, add one to the counter
		r.Wg.Add(1)
		go func(db operatorkubecentercomv1beta1.BackUp) {
			//decrement the counter when the task is done
			defer r.Wg.Done()
			for {
				// time is up for the next ticker, start the backup task
				<-ticker.C
				// reset the ticker. since the task is executed periodically
				ticker.Reset(time.Duration(db.Spec.Period) * time.Minute)
				// TODO: execute the task

			}
		}(backup)
	}

	// wait for the go routines
	r.Wg.Wait()

}

func (r *BackUpReconciler) GetActualStartTime(startTime string) time.Duration {
	// get the expected start time from request
	hoursAndMinutes := strings.Split(startTime, ":")
	hours, _ := strconv.Atoi(hoursAndMinutes[0])
	minutes, _ := strconv.Atoi(hoursAndMinutes[1])

	now := time.Now().Truncate(time.Second)

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(time.Hour * 24)

	// the result to be returned should be the duration
	var seconds int

	expectedTime := time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes)
	currentTime := time.Hour*time.Duration(now.Hour()) + time.Minute*time.Duration(now.Minute())

	if currentTime >= expectedTime {
		// executed it after 24 hours
		seconds = int(todayEnd.Add(expectedTime).Sub(now).Seconds())
	} else {
		seconds = int(todayStart.Add(expectedTime).Sub(now).Seconds())
	}
	return time.Duration(seconds) * time.Second
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackUpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorkubecentercomv1beta1.BackUp{}).
		Complete(r)
}

func (r *BackUpReconciler) GetNextTime(currentSeconds float64) time.Time {
	currentTime := time.Now()
	return currentTime.Add(time.Duration(currentSeconds) * time.Second)
}

func (r *BackUpReconciler) UpdateStatus(backup operatorkubecentercomv1beta1.BackUp) {
	r.lock.Lock()
	defer r.lock.Unlock()
	ctx := context.TODO()

	namespacedName := types.NamespacedName{
		Name:      backup.Name,
		Namespace: backup.Namespace,
	}

	var backupK8S operatorkubecentercomv1beta1.BackUp

	err := r.Get(ctx, namespacedName, &backupK8S)
	if err != nil {
		return
	}
	// update the backup status, including all the fields in backup.Status
	backupK8S.Status = backup.Status
	err = r.Client.Status().Update(ctx, &backupK8S)
	if err != nil {
		return
	}
}
