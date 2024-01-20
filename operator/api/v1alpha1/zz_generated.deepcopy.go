//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023 George Onoufriou.

Licensed under the Open Software Licence, Version 3.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License in the project root (LICENSE) or at

    https://opensource.org/license/osl-3-0-php/
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"encoding/json"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ak) DeepCopyInto(out *Ak) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ak.
func (in *Ak) DeepCopy() *Ak {
	if in == nil {
		return nil
	}
	out := new(Ak)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Ak) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkBlueprint) DeepCopyInto(out *AkBlueprint) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkBlueprint.
func (in *AkBlueprint) DeepCopy() *AkBlueprint {
	if in == nil {
		return nil
	}
	out := new(AkBlueprint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AkBlueprint) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkBlueprintList) DeepCopyInto(out *AkBlueprintList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AkBlueprint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkBlueprintList.
func (in *AkBlueprintList) DeepCopy() *AkBlueprintList {
	if in == nil {
		return nil
	}
	out := new(AkBlueprintList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AkBlueprintList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkBlueprintSpec) DeepCopyInto(out *AkBlueprintSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkBlueprintSpec.
func (in *AkBlueprintSpec) DeepCopy() *AkBlueprintSpec {
	if in == nil {
		return nil
	}
	out := new(AkBlueprintSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkBlueprintStatus) DeepCopyInto(out *AkBlueprintStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkBlueprintStatus.
func (in *AkBlueprintStatus) DeepCopy() *AkBlueprintStatus {
	if in == nil {
		return nil
	}
	out := new(AkBlueprintStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkList) DeepCopyInto(out *AkList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Ak, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkList.
func (in *AkList) DeepCopy() *AkList {
	if in == nil {
		return nil
	}
	out := new(AkList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AkList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkSpec) DeepCopyInto(out *AkSpec) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make(json.RawMessage, len(*in))
		copy(*out, *in)
	}
	if in.Blueprints != nil {
		in, out := &in.Blueprints, &out.Blueprints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkSpec.
func (in *AkSpec) DeepCopy() *AkSpec {
	if in == nil {
		return nil
	}
	out := new(AkSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AkStatus) DeepCopyInto(out *AkStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AkStatus.
func (in *AkStatus) DeepCopy() *AkStatus {
	if in == nil {
		return nil
	}
	out := new(AkStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthentikInstance) DeepCopyInto(out *AuthentikInstance) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthentikInstance.
func (in *AuthentikInstance) DeepCopy() *AuthentikInstance {
	if in == nil {
		return nil
	}
	out := new(AuthentikInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BP) DeepCopyInto(out *BP) {
	*out = *in
	in.Metadata.DeepCopyInto(&out.Metadata)
	if in.Context != nil {
		in, out := &in.Context, &out.Context
		*out = (*in).DeepCopy()
	}
	if in.Entries != nil {
		in, out := &in.Entries, &out.Entries
		*out = make([]BPModel, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BP.
func (in *BP) DeepCopy() *BP {
	if in == nil {
		return nil
	}
	out := new(BP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BPMeta) DeepCopyInto(out *BPMeta) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BPMeta.
func (in *BPMeta) DeepCopy() *BPMeta {
	if in == nil {
		return nil
	}
	out := new(BPMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BPModel) DeepCopyInto(out *BPModel) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Identifiers != nil {
		in, out := &in.Identifiers, &out.Identifiers
		*out = (*in).DeepCopy()
	}
	if in.Attrs != nil {
		in, out := &in.Attrs, &out.Attrs
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BPModel.
func (in *BPModel) DeepCopy() *BPModel {
	if in == nil {
		return nil
	}
	out := new(BPModel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigmapSettings) DeepCopyInto(out *ConfigmapSettings) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigmapSettings.
func (in *ConfigmapSettings) DeepCopy() *ConfigmapSettings {
	if in == nil {
		return nil
	}
	out := new(ConfigmapSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDC) DeepCopyInto(out *OIDC) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDC.
func (in *OIDC) DeepCopy() *OIDC {
	if in == nil {
		return nil
	}
	out := new(OIDC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDC) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCApplication) DeepCopyInto(out *OIDCApplication) {
	*out = *in
	out.ConfigMap = in.ConfigMap
	if in.BackChannelProviders != nil {
		in, out := &in.BackChannelProviders, &out.BackChannelProviders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.OIDCApplicationUISettings = in.OIDCApplicationUISettings
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCApplication.
func (in *OIDCApplication) DeepCopy() *OIDCApplication {
	if in == nil {
		return nil
	}
	out := new(OIDCApplication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCApplicationUISettings) DeepCopyInto(out *OIDCApplicationUISettings) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCApplicationUISettings.
func (in *OIDCApplicationUISettings) DeepCopy() *OIDCApplicationUISettings {
	if in == nil {
		return nil
	}
	out := new(OIDCApplicationUISettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCList) DeepCopyInto(out *OIDCList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OIDC, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCList.
func (in *OIDCList) DeepCopy() *OIDCList {
	if in == nil {
		return nil
	}
	out := new(OIDCList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDCList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCProvider) DeepCopyInto(out *OIDCProvider) {
	*out = *in
	out.Secret = in.Secret
	in.ProtocolSettings.DeepCopyInto(&out.ProtocolSettings)
	in.MachineToMachineSettings.DeepCopyInto(&out.MachineToMachineSettings)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCProvider.
func (in *OIDCProvider) DeepCopy() *OIDCProvider {
	if in == nil {
		return nil
	}
	out := new(OIDCProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCProviderMachineToMachineSettings) DeepCopyInto(out *OIDCProviderMachineToMachineSettings) {
	*out = *in
	if in.TrustedOIDCSources != nil {
		in, out := &in.TrustedOIDCSources, &out.TrustedOIDCSources
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCProviderMachineToMachineSettings.
func (in *OIDCProviderMachineToMachineSettings) DeepCopy() *OIDCProviderMachineToMachineSettings {
	if in == nil {
		return nil
	}
	out := new(OIDCProviderMachineToMachineSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCProviderProtocolSettings) DeepCopyInto(out *OIDCProviderProtocolSettings) {
	*out = *in
	if in.RedirectURIs != nil {
		in, out := &in.RedirectURIs, &out.RedirectURIs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCProviderProtocolSettings.
func (in *OIDCProviderProtocolSettings) DeepCopy() *OIDCProviderProtocolSettings {
	if in == nil {
		return nil
	}
	out := new(OIDCProviderProtocolSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCSpec) DeepCopyInto(out *OIDCSpec) {
	*out = *in
	out.Instance = in.Instance
	if in.Providers != nil {
		in, out := &in.Providers, &out.Providers
		*out = make([]OIDCProvider, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Applications != nil {
		in, out := &in.Applications, &out.Applications
		*out = make([]OIDCApplication, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCSpec.
func (in *OIDCSpec) DeepCopy() *OIDCSpec {
	if in == nil {
		return nil
	}
	out := new(OIDCSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCStatus) DeepCopyInto(out *OIDCStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCStatus.
func (in *OIDCStatus) DeepCopy() *OIDCStatus {
	if in == nil {
		return nil
	}
	out := new(OIDCStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretSettings) DeepCopyInto(out *SecretSettings) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretSettings.
func (in *SecretSettings) DeepCopy() *SecretSettings {
	if in == nil {
		return nil
	}
	out := new(SecretSettings)
	in.DeepCopyInto(out)
	return out
}
