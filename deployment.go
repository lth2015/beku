package beku

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/yulibaozi/mapper"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment include Kubernetes resource object Deployment and error
type Deployment struct {
	dp  *v1.Deployment
	err error
}

// NewDeployment create Deployment and Chain function call begin with this function.
func NewDeployment() *Deployment { return &Deployment{dp: &v1.Deployment{}} }

// Finish Chain function call end with this function
// return Kubernetes resource object Deployment and error.
// In the function, it will check necessary parameters、input the default field。
func (obj *Deployment) Finish() (dp *v1.Deployment, err error) {
	obj.verify()
	dp, err = obj.dp, obj.err
	return
}

// JSONNew use json data create Deployment
func (obj *Deployment) JSONNew(jsonbyts []byte) *Deployment {
	obj.err = json.Unmarshal(jsonbyts, obj.dp)
	return obj
}

// YAMLNew use yaml data create Deployment
func (obj *Deployment) YAMLNew(yamlbyts []byte) *Deployment {
	obj.err = yaml.Unmarshal(yamlbyts, obj.dp)
	return obj
}

// SetName set Deployment name
func (obj *Deployment) SetName(name string) *Deployment {
	obj.dp.SetName(name)
	return obj
}

// SetNamespace set Deployment namespace and set Pod namespace.
func (obj *Deployment) SetNamespace(namespace string) *Deployment {
	obj.dp.SetNamespace(namespace)
	obj.dp.Spec.Template.SetNamespace(namespace)
	return obj
}

// SetNamespaceAndName set Deployment namespace,set Pod namespace,set Deployment name.
func (obj *Deployment) SetNamespaceAndName(namespace, name string) *Deployment {
	obj.SetNamespace(namespace)
	obj.SetName(name)
	return obj
}

// SetLabels set Deployment labels and set Pod Labels
func (obj *Deployment) SetLabels(labels map[string]string) *Deployment {
	obj.dp.SetLabels(labels)
	obj.dp.Spec.Template.SetLabels(labels)
	return obj
}

// SetSelector set Deployment selector
// set:
// 1. Deployment.Spec.Selector
// 2. Deployment.Spec.Template.Label(the Field is Pod Labels.)
// and you can not be SetLabels
func (obj *Deployment) SetSelector(labels map[string]string) *Deployment {
	if len(labels) <= 0 {
		obj.err = errors.New("LabelSelector set error,Labels is not allowed to be empty ")
		return obj
	}
	if obj.dp.Spec.Selector == nil {
		obj.dp.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}
		obj.SetLabels(labels)
		return obj
	}
	obj.SetLabels(labels)
	obj.dp.Spec.Selector.MatchLabels = labels
	return obj
}

// SetAnnotations set Deployment annotations
func (obj *Deployment) SetAnnotations(annotations map[string]string) *Deployment {
	obj.dp.SetAnnotations(annotations)
	return obj
}

// SetReplicas set Deployment replicas default 1
func (obj *Deployment) SetReplicas(replicas int32) *Deployment {
	obj.dp.Spec.Replicas = &replicas
	return obj
}

// SetMinReadySeconds set Deployment minreadyseconds default 600
func (obj *Deployment) SetMinReadySeconds(sec int32) *Deployment {
	if sec < 0 {
		sec = 0
	}
	obj.dp.Spec.MinReadySeconds = sec
	return obj
}

// SetHistoryLimit set Deployment history version numbers, limit default 10
// the field is used to Rollback
func (obj *Deployment) SetHistoryLimit(limit int32) *Deployment {
	if limit <= 0 {
		limit = 10
	}
	obj.dp.Spec.RevisionHistoryLimit = &limit
	return obj
}

// SetHTTPLiveness set container liveness of http style
// port: required
// path: http request URL,eg: /api/v1/posts/1
// initDelaySec: how long time after the first start of the program the probe is executed for the first time.(sec)
// timeoutSec: http request timeout seconds,defaults to 1 second. Minimum value is 1.
// periodSec: how often does the probe??defaults to 1 second. Minimum value is 1,Except for the first time?
// headers: headers[0] is HTTP Header, do not fill if you do not need to set
// on the other hand, only **first container** will be set livenessProbe
func (obj *Deployment) SetHTTPLiveness(port int, path string, initDelaySec, timeoutSec, periodSec int32, headers ...map[string]string) *Deployment {
	return obj.setLiveness(httpProbe(port, path, initDelaySec, timeoutSec, periodSec, headers...))
}

// SetCMDLiveness set container liveness of cmd style
// cmd: execute liveness probe as commond line
// timeoutSec: http request timeout seconds,defaults to 1 second. Minimum value is 1.
// periodSec: how often does the probe??defaults to 1 second. Minimum value is 1,Except for the first time?
// headers: headers[0] is HTTP Header, do not fill if you do not need to set
// on the other hand, only **first container** will be set livenessProbe
func (obj *Deployment) SetCMDLiveness(cmd []string, initDelaySec, timeoutSec, periodSec int32) *Deployment {
	return obj.setLiveness(cmdProbe(cmd, initDelaySec, timeoutSec, periodSec))
}

// SetTCPLiveness set container liveness of tcp style
// host: default is ""
// port: required
// timeoutSec: http request timeout seconds,defaults to 1 second. Minimum value is 1.
// periodSec: how often does the probe??defaults to 1 second. Minimum value is 1,Except for the first time?
// headers: headers[0] is HTTP Header, do not fill if you do not need to set
// on the other hand, only **first container** will be set livenessProbe
func (obj *Deployment) SetTCPLiveness(host string, port int, initDelaySec, timeoutSec, periodSec int32) *Deployment {
	return obj.setLiveness(tcpProbe(host, port, initDelaySec, timeoutSec, periodSec))
}

func (obj *Deployment) setLiveness(probe *corev1.Probe) *Deployment {
	if len(obj.dp.Spec.Template.Spec.Containers) <= 0 {
		obj.dp.Spec.Template.Spec.Containers = []corev1.Container{corev1.Container{LivenessProbe: probe}}
		return obj
	}
	obj.dp.Spec.Template.Spec.Containers[0].LivenessProbe = probe
	return obj
}

func (obj *Deployment) setReadness(probe *corev1.Probe) *Deployment {
	if len(obj.dp.Spec.Template.Spec.Containers) <= 0 {
		obj.dp.Spec.Template.Spec.Containers = []corev1.Container{corev1.Container{ReadinessProbe: probe}}
		return obj
	}
	obj.dp.Spec.Template.Spec.Containers[0].ReadinessProbe = probe
	return obj
}

// SetHTTPReadness set container readness
// initDelaySec: how long time after the first start of the program the probe is executed for the first time.(sec)
// timeoutSec: http request timeout seconds,defaults to 1 second. Minimum value is 1.
// periodSec: how often does the probe??defaults to 1 second. Minimum value is 1,Except for the first time?
// on the other hand, only **first container** will be set livenessProbe
func (obj *Deployment) SetHTTPReadness(port int, path string, initDelaySec, timeoutSec, periodSec int32, headers ...map[string]string) *Deployment {
	return obj.setReadness(httpProbe(port, path, initDelaySec, timeoutSec, periodSec, headers...))
}

// SetCMDReadness set container readness of cmd style
// cmd: execute readness probe as commond line
// timeoutSec: http request timeout seconds,defaults to 1 second. Minimum value is 1.
// periodSec: how often does the probe? defaults to 1 second. Minimum value is 1,Except for the first time?
// headers: headers[0] is HTTP Header, do not fill if you do not need to set
// on the other hand, only **first container** will be set livenessProbe
func (obj *Deployment) SetCMDReadness(cmd []string, initDelaySec, timeoutSec, periodSec int32) *Deployment {
	return obj.setReadness(cmdProbe(cmd, initDelaySec, timeoutSec, periodSec))
}

// SetTCPReadness set container readness of tcp style
// host: default is ""
// port: required
// timeoutSec: http request timeout seconds,defaults to 1 second. Minimum value is 1.
// periodSec: how often does the probe? defaults to 1 second. Minimum value is 1,Except for the first time?
// headers: headers[0] is HTTP Header, do not fill if you do not need to set
// on the other hand, only **first container** will be set livenessProbe
func (obj *Deployment) SetTCPReadness(host string, port int, initDelaySec, timeoutSec, periodSec int32) *Deployment {
	return obj.setReadness(tcpProbe(host, port, initDelaySec, timeoutSec, periodSec))
}

// SetMatchExpressions set Deployment match expressions
// the field is used to set complicated Label.
// ToDo: mapper error.
func (obj *Deployment) SetMatchExpressions(ents []LabelSelectorRequirement) *Deployment {
	requirements := make([]metav1.LabelSelectorRequirement, 0)
	err := mapper.AutoMapper(ents, requirements)
	if err != nil {
		obj.err = fmt.Errorf("SetMatchExpressions err:%v", err)
		return obj
	}
	if obj.dp.Spec.Selector == nil {
		obj.dp.Spec.Selector = &metav1.LabelSelector{
			MatchExpressions: requirements,
		}
		return obj
	}
	obj.dp.Spec.Selector.MatchExpressions = requirements
	return obj
}

// SetDeployMaxTime set Deployment deploy max time,default 600s.
// If real deploy time more than this value,Deployment controller return err:ProgressDeadlineExceeded
// and Pod will Redeploy.
func (obj *Deployment) SetDeployMaxTime(sec int32) *Deployment {
	if sec < 0 {
		sec = 600
	}
	obj.dp.Spec.ProgressDeadlineSeconds = &sec
	return obj
}

// SetPodLabels set Pod labels
// when call SetLabels(),you can not use this function.
func (obj *Deployment) SetPodLabels(labels map[string]string) *Deployment {
	obj.dp.Spec.Template.SetLabels(labels)
	return obj
}

// GetPodLabel get Pod labels
func (obj *Deployment) GetPodLabel() map[string]string {
	return obj.dp.Spec.Template.GetLabels()
}

// SetPVClaim set Deployment PersistentVolumeClaimVolumeSource
// params:
// volumeName: this is Custom field,you can define VolumeSource name,will be used of the container MountPath,
// claimName: this is PersistentVolumeClaim(PVC) name,the PVC and Deployment must on same namespace and exist.
func (obj *Deployment) SetPVClaim(volumeName, claimName string) *Deployment {
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
				ReadOnly:  false,
			},
		},
	}
	if len(obj.dp.Spec.Template.Spec.Volumes) <= 0 {
		obj.dp.Spec.Template.Spec.Volumes = []corev1.Volume{volume}
		return obj
	}
	obj.dp.Spec.Template.Spec.Volumes = append(obj.dp.Spec.Template.Spec.Volumes, volume)
	return obj
}

//SetPVCMounts mount PersistentVolumeClaim on container
// params:
// volumeName:the param is SetPVClaim() function volumeName,and when you call SetPVCMounts function you must call SetPVClaim function,and no order.
// on the other hand SetPVCMounts() function only mount first Container,and On the Container you can volumeMount many PersistentVolumeClaim.
// mounthPath: runtime container dir eg:/var/lib/mysql
func (obj *Deployment) SetPVCMounts(volumeName, mounthPath string) *Deployment {
	volumeMount := corev1.VolumeMount{Name: volumeName, MountPath: mounthPath}
	if len(obj.dp.Spec.Template.Spec.Containers) <= 0 {
		obj.dp.Spec.Template.Spec.Containers = append(obj.dp.Spec.Template.Spec.Containers, corev1.Container{
			VolumeMounts: []corev1.VolumeMount{volumeMount},
		})
		return obj
	}
	//only mount first container and first container can mount many data source.
	if len(obj.dp.Spec.Template.Spec.Containers[0].VolumeMounts) <= 0 {
		obj.dp.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{volumeMount}
		return obj
	}
	obj.dp.Spec.Template.Spec.Containers[0].VolumeMounts = append(obj.dp.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)
	return obj
}

// SetContainer set Deployment container
// name:name is container name ,default ""
// image:image is image name ,must input image
// containerPort: image expose containerPort,must input containerPort
func (obj *Deployment) SetContainer(name, image string, containerPort int32) *Deployment {
	// This must be a valid port number, 0 < x < 65536.
	if containerPort <= 0 || containerPort >= 65536 {
		obj.err = errors.New("SetContainer err, container Port range: 0 < containerPort < 65536")
		return obj
	}
	if !verifyString(image) {
		obj.err = errors.New("SetContainer err, image is not allowed to be empty")
		return obj
	}
	port := corev1.ContainerPort{ContainerPort: containerPort}
	container := corev1.Container{
		Name:  name,
		Image: image,
		Ports: []corev1.ContainerPort{port},
	}
	containersLen := len(obj.dp.Spec.Template.Spec.Containers)
	if containersLen < 1 {
		obj.dp.Spec.Template.Spec.Containers = []corev1.Container{container}
		return obj
	}
	for index := 0; index < containersLen; index++ {
		img := strings.TrimSpace(obj.dp.Spec.Template.Spec.Containers[index].Image)
		if img == "" || len(img) <= 0 {
			obj.dp.Spec.Template.Spec.Containers[index].Name = name
			obj.dp.Spec.Template.Spec.Containers[index].Image = image
			obj.dp.Spec.Template.Spec.Containers[index].Ports = []corev1.ContainerPort{port}
			return obj
		}
	}
	obj.dp.Spec.Template.Spec.Containers = append(obj.dp.Spec.Template.Spec.Containers, container)
	return obj
}

// SetEnvs set Pod Environmental variable
func (obj *Deployment) SetEnvs(envMap map[string]string) *Deployment {
	envs, err := mapToEnvs(envMap)
	if err != nil {
		obj.err = err
		return obj
	}
	containerLen := len(obj.dp.Spec.Template.Spec.Containers)
	if containerLen < 1 {
		obj.dp.Spec.Template.Spec.Containers = []corev1.Container{corev1.Container{Env: envs}}
		return obj
	}
	for index := 0; index < containerLen; index++ {
		if obj.dp.Spec.Template.Spec.Containers[index].Env == nil {
			obj.dp.Spec.Template.Spec.Containers[index].Env = envs
		}
	}
	return obj
}

// verify check service necessary value, input the default field and input related data.
func (obj *Deployment) verify() {
	if obj.err != nil {
		return
	}
	if !verifyString(obj.dp.GetName()) {
		obj.err = errors.New("Deployment name is not allowed to be empty")
		return
	}
	if len(obj.dp.GetLabels()) < 1 {
		obj.err = errors.New("Deployment Labels is not allowed to be empty")
		return
	}
	if len(obj.dp.Spec.Template.GetLabels()) < 1 {
		obj.err = errors.New("Deployment.Spec.Templata.Labels is not allowed to be empty")
		return
	}
	if obj.dp.Spec.Template.Spec.Containers == nil || len(obj.dp.Spec.Template.Spec.Containers) < 1 {
		obj.err = errors.New("Deployment.Spec.Template.Spec.Containers is not allowed to be empty")
		return
	}
	if obj.dp.Spec.Selector == nil {
		obj.SetSelector(obj.GetPodLabel())
	}
	obj.dp.Kind = "Deployment"
	obj.dp.APIVersion = "apps/v1"
	return
}
