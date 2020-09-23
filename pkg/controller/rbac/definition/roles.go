/*
Copyright 2020 The Crossplane Authors.

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

package definition

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane/apis/apiextensions/v1alpha1"
)

const (
	namePrefix       = "crossplane:composite:"
	nameSuffixEdit   = ":aggregate-to-edit"
	nameSuffixView   = ":aggregate-to-view"
	nameSuffixBrowse = ":aggregate-to-browse"

	keyAggregateToCrossplane = "rbac.crossplane.io/aggregate-to-crossplane"

	keyAggregateToAdmin   = "rbac.crossplane.io/aggregate-to-admin"
	keyAggregateToNSAdmin = "rbac.crossplane.io/aggregate-to-ns-admin"

	keyAggregateToEdit   = "rbac.crossplane.io/aggregate-to-edit"
	keyAggregateToNSEdit = "rbac.crossplane.io/aggregate-to-ns-edit"

	keyAggregateToView   = "rbac.crossplane.io/aggregate-to-view"
	keyAggregateToNSView = "rbac.crossplane.io/aggregate-to-ns-view"

	keyAggregateToBrowse = "rbac.crossplane.io/aggregate-to-browse"

	keyXRD = "rbac.crossplane.io/xrd"

	valTrue = "true"
)

var (
	verbsEdit   = []string{rbacv1.VerbAll}
	verbsView   = []string{"get", "list", "watch"}
	verbsBrowse = []string{"get", "list", "watch"}
)

// RenderClusterRoles returns ClusterRoles for the supplied XRD.
func RenderClusterRoles(d *v1alpha1.CompositeResourceDefinition) []rbacv1.ClusterRole {
	edit := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: namePrefix + d.GetName() + nameSuffixEdit,
			Labels: map[string]string{
				// Edit rules aggregate to the Crossplane ClusterRole too.
				// Crossplane needs access to reconcile all composite resources
				// and composite resource claims.
				keyAggregateToCrossplane: valTrue,

				// Edit rules aggregate to admin too. Currently edit and admin
				// differ only in their base roles.
				keyAggregateToAdmin:   valTrue,
				keyAggregateToNSAdmin: valTrue,

				keyAggregateToEdit:   valTrue,
				keyAggregateToNSEdit: valTrue,

				keyXRD: d.GetName(),
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{d.Spec.CRDSpecTemplate.Group},
				Resources: []string{d.Spec.CRDSpecTemplate.Names.Plural},
				Verbs:     verbsEdit,
			},
		},
	}

	view := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: namePrefix + d.GetName() + nameSuffixView,
			Labels: map[string]string{
				keyAggregateToView:   valTrue,
				keyAggregateToNSView: valTrue,

				keyXRD: d.GetName(),
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{d.Spec.CRDSpecTemplate.Group},
				Resources: []string{d.Spec.CRDSpecTemplate.Names.Plural},
				Verbs:     verbsView,
			},
		},
	}

	browse := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: namePrefix + d.GetName() + nameSuffixBrowse,
			Labels: map[string]string{
				keyAggregateToBrowse: valTrue,

				keyXRD: d.GetName(),
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{d.Spec.CRDSpecTemplate.Group},
				Resources: []string{d.Spec.CRDSpecTemplate.Names.Plural},
				Verbs:     verbsBrowse,
			},
		},
	}

	if d.Spec.ClaimNames != nil {
		edit.Rules = append(edit.Rules, rbacv1.PolicyRule{
			APIGroups: []string{d.Spec.CRDSpecTemplate.Group},
			Resources: []string{d.Spec.ClaimNames.Plural},
			Verbs:     verbsEdit,
		})

		view.Rules = append(view.Rules, rbacv1.PolicyRule{
			APIGroups: []string{d.Spec.CRDSpecTemplate.Group},
			Resources: []string{d.Spec.ClaimNames.Plural},
			Verbs:     verbsView,
		})

		// The browse role only includes composite resources; not claims.
	}

	for _, o := range []metav1.Object{edit, view, browse} {
		meta.AddOwnerReference(o, meta.AsController(meta.TypedReferenceTo(d, v1alpha1.CompositeResourceDefinitionGroupVersionKind)))
	}

	return []rbacv1.ClusterRole{*edit, *view, *browse}
}