package handler

import (
	"fmt"

	apiV1 "k8s.io/api/core/v1"
	"k8sync/internal/k8s/utils"
)

// Event represent an event got from k8s api server
// Events from different endpoints need to be cast to controller
// before being able to be handled by handler
type Event struct {
	Namespace string
	Kind      string
	Component string
	Host      string
	Reason    string
	Status    string
	Name      string
	Obj       interface{}
}

// New create new handler event
func New(obj interface{}, ns string, action string, rtype string, status string) *Event {
	var namespace, kind, component, host, name, reason string

	objectMeta := utils.GetObjectMetaData(obj)
	//namespace = objectMeta.Namespace
	namespace = ns
	name = objectMeta.Name
	reason = action

	switch object := obj.(type) {
	case *apiV1.Pod:
		//kind = "pod"
		host = object.Spec.NodeName
	case *apiV1.Service:
		//kind = "service"
		component = string(object.Spec.Type)
	case Event:
		name = object.Name
		//kind = object.Kind
		namespace = object.Namespace
	}
	kind = rtype

	kbEvent := Event{
		Namespace: namespace,
		Kind:      kind,
		Component: component,
		Host:      host,
		Reason:    reason,
		Status:    status,
		Name:      name,
		Obj:       obj,
	}
	return &kbEvent
}

// Message returns event message in standard format.
// included as a part of event packege to enhance code resuablity across handlers.
func (e *Event) Message() (msg string) {
	// using switch over if..else, since the format could vary based on the kind of the object in future.
	switch e.Kind {
	case "namespace":
		msg = fmt.Sprintf(
			"A namespace `%s` has been `%s`",
			e.Name,
			e.Reason,
		)
	case "node":
		msg = fmt.Sprintf(
			"A node `%s` has been `%s`",
			e.Name,
			e.Reason,
		)
	case "cluster role":
		msg = fmt.Sprintf(
			"A cluster role `%s` has been `%s`",
			e.Name,
			e.Reason,
		)
	case "NodeReady":
		msg = fmt.Sprintf(
			"Node `%s` is Ready : \nNodeReady",
			e.Name,
		)
	case "NodeNotReady":
		msg = fmt.Sprintf(
			"Node `%s` is Not Ready : \nNodeNotReady",
			e.Name,
		)
	case "NodeRebooted":
		msg = fmt.Sprintf(
			"Node `%s` Rebooted : \nNodeRebooted",
			e.Name,
		)
	case "Backoff":
		msg = fmt.Sprintf(
			"Pod `%s` in `%s` Crashed : \nCrashLoopBackOff %s",
			e.Name,
			e.Namespace,
			e.Reason,
		)
	default:
		msg = fmt.Sprintf(
			"A `%s` in namespace `%s` has been `%s`:\n`%s`",
			e.Kind,
			e.Namespace,
			e.Reason,
			e.Name,
		)
	}
	return msg
}
