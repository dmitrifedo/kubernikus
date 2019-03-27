package servicing

import (
	"time"

	"k8s.io/client-go/tools/record"

	"github.com/pkg/errors"

	"github.com/sapcc/kubernikus/pkg/controller/nodeobservatory"

	"github.com/go-kit/kit/log"

	"github.com/sapcc/kubernikus/pkg/api/models"
	client "github.com/sapcc/kubernikus/pkg/generated/clientset/typed/kubernikus/v1"
	listers_kubernikus_v1 "github.com/sapcc/kubernikus/pkg/generated/listers/kubernikus/v1"
	"github.com/sapcc/kubernikus/pkg/util"

	v1 "github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
)

const (
	// AnnotationServicingTimestamp holds the timestamp when the kluster was last serviced
	AnnotationServicingTimestamp = "kubernikus.cloud.sap/servicingTimestamp"
)

var (
	// ServiceInterval defines how often a kluster is serviced
	ServiceInterval = 1 * time.Hour
)

type (
	// Reconciler checks a specific kluster's nodes for updates/upgrades
	Reconciler interface {
		Do() error
	}

	// ReconcilerFactory produces a Reconciler
	ReconcilerFactory interface {
		Make(*v1.Kluster) (Reconciler, error)
	}

	// KlusterReconcilerFactory produces a Reconciler for a specific kluster
	KlusterReconcilerFactory struct {
		Logger            log.Logger
		ListerFactory     ListerFactory
		LifeCyclerFactory LifeCyclerFactory
		KlusterLister     listers_kubernikus_v1.KlusterLister
		KubernikusClient  client.KubernikusV1Interface
	}

	// KlusterReconciler is a concrete implementation of a Reconciler
	KlusterReconciler struct {
		Logger           log.Logger
		Kluster          *v1.Kluster
		Lister           Lister
		LifeCycler       LifeCycler
		KlusterLister    listers_kubernikus_v1.KlusterLister
		KubernikusClient client.KubernikusV1Interface
	}

	// LoggingReconciler decorates a Reconciler with log messages
	LoggingReconciler struct {
		Logger     log.Logger
		Reconciler Reconciler
	}
)

// NewKlusterReconcilerFactory produces a new Factory
func NewKlusterReconcilerFactory(logger log.Logger, recorder record.EventRecorder, no *nodeobservatory.NodeObservatory,
	kl listers_kubernikus_v1.KlusterLister, kc client.KubernikusV1Interface) *KlusterReconcilerFactory {
	return &KlusterReconcilerFactory{
		Logger: logger,
		ListerFactory: &NodeListerFactory{
			Logger:          logger,
			NodeObservatory: no,
			CoreOSVersion:   &LatestCoreOSVersion{},
		},
		LifeCyclerFactory: &NodeLifeCyclerFactory{
			Recorder: recorder,
			Logger:   logger,
		},
		KlusterLister:    kl,
		KubernikusClient: kc,
	}
}

// Make a new Factory
func (f *KlusterReconcilerFactory) Make(k *v1.Kluster) (Reconciler, error) {
	logger := log.With(f.Logger, "kluster", k.Spec.Name, "project", k.Account())
	lister, err := f.ListerFactory.Make(k)
	if err != nil {
		return nil, err
	}

	cycler, err := f.LifeCyclerFactory.Make(k)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't make LifeCyclerFactory")
	}

	var reconciler Reconciler
	reconciler = &KlusterReconciler{
		Logger:           logger,
		Kluster:          k,
		Lister:           lister,
		LifeCycler:       cycler,
		KlusterLister:    f.KlusterLister,
		KubernikusClient: f.KubernikusClient,
	}

	reconciler = &LoggingReconciler{
		Logger:     logger,
		Reconciler: reconciler,
	}

	return reconciler, nil
}

// Do it
func (r *KlusterReconciler) Do() error {
	r.Logger.Log("msg", "reconciling", "v", 2)

	if r.Kluster.Status.Phase != models.KlusterPhaseRunning {
		r.Logger.Log("msg", "skipped upgrades because kluster is not running", "v", 2)
		return nil
	}

	if !r.isServiceIntervalElapsed() {
		r.Logger.Log("msg", "skipped upgrades because kluster service interval not elapsed yet", "v", 2)
		return nil
	}

	notReady := r.Lister.NotReady()
	requiringReboot := r.Lister.RequiringReboot()
	requiringReplacement := r.Lister.RequiringReplacement()

	if len(notReady) > 0 {
		r.Logger.Log("msg", "skipped upgrades because kluster is not healthy", "v", 2)
		return nil
	}

	if len(requiringReplacement) > 0 {
		if err := r.LifeCycler.Drain(requiringReplacement[0]); err != nil {
			return errors.Wrap(err, "Failed to drain node that is about to be replaces")
		}

		if err := r.LifeCycler.Replace(requiringReplacement[0]); err != nil {
			return errors.Wrap(err, "Failed to replace node")
		}
	} else if len(requiringReboot) > 0 {
		if err := r.LifeCycler.Drain(requiringReboot[0]); err != nil {
			return errors.Wrap(err, "Failed to drain node that is about to be rebooted")
		}

		if err := r.LifeCycler.Reboot(requiringReboot[0]); err != nil {
			return errors.Wrap(err, "Failed to reboot node")
		}
	}

	return r.updateLastServicingTime()
}

// Do log it
func (r *LoggingReconciler) Do() (err error) {
	defer func(begin time.Time) {
		r.Logger.Log(
			"msg", "reconciled",
			"took", time.Since(begin),
			"v", 2,
			"err", err,
		)
	}(time.Now())
	return r.Reconciler.Do()
}

func (r *KlusterReconciler) updateLastServicingTime() error {
	client := r.KubernikusClient.Klusters(r.Kluster.Namespace)
	lister := r.KlusterLister.Klusters(r.Kluster.Namespace)
	_, err := util.UpdateKlusterWithRetries(client, lister, r.Kluster.Name, func(kluster *v1.Kluster) error {
		kluster.ObjectMeta.Annotations[AnnotationServicingTimestamp] = Now().UTC().Format(time.RFC3339)
		return nil
	})
	return err
}

func (r *KlusterReconciler) getLastServicingTime(annotations map[string]string) time.Time {
	t, ok := annotations[AnnotationServicingTimestamp]
	if !ok {
		return time.Unix(0, 0)
	}

	pt, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Unix(0, 0)
	}

	return pt
}

func (r *KlusterReconciler) isServiceIntervalElapsed() bool {
	nextServiceTime := r.getLastServicingTime(r.Kluster.ObjectMeta.GetAnnotations()).Add(ServiceInterval)
	return Now().After(nextServiceTime)
}
