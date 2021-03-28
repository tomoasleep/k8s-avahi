package controller

import (
	"context"
	"strings"

	"net"

	"github.com/go-logr/logr"
	"github.com/tomoasleep/k8s-avahi/mdns"
	networkingv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	client.Client
	Log        logr.Logger
	MdnsClient *mdns.MdnsClient
}

type Target struct {
	ip       net.IP
	hostname string
}

var (
	// TLD is a domain to manage with Avahi
	TLD = "local"
)

// SetupManager controller routine
func (c *Controller) SetupManager(mgr ctrl.Manager) error {
	_, err := ctrl.NewControllerManagedBy(mgr).For(&networkingv1.Ingress{}).Build(c)
	if err != nil {
		return err
	}

	return nil
}

// Close removes created mdns records
func (c *Controller) Close() error {
	c.MdnsClient.Close()
	return nil
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ingress := &networkingv1.Ingress{}
	c.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, ingress)

	targets := targetsFromIngressStatus(ingress.Status)

	for _, rule := range ingress.Spec.Rules {
		if rule.Host == "" {
			continue
		}

		names := strings.Split(rule.Host, ".")
		if len(names) != 2 {
			continue
		}

		if names[1] != TLD {
			continue
		}

		err := c.reconcileHost(names[0], targets)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (c *Controller) reconcileHost(hostname string, targets []Target) error {
	for _, target := range targets {
		if target.ip != nil {
			return c.MdnsClient.RegisterRecord(strings.Join([]string{hostname, TLD}, "."), target.ip)
		}
	}

	return nil
}

func targetsFromIngressStatus(status networkingv1.IngressStatus) []Target {
	targets := []Target{}

	for _, lb := range status.LoadBalancer.Ingress {
		if lb.IP != "" {
			ip := net.ParseIP(lb.IP)
			targets = append(targets, Target{ip: ip})
		}
		if lb.Hostname != "" {
			targets = append(targets, Target{hostname: lb.Hostname})
		}
	}

	return targets
}
