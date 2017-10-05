// +build !ignore_autogenerated

/*
Copyright 2017 The Kubernetes Authors.

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package testing

import (
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	reflect "reflect"
)

// Deprecated: GetGeneratedDeepCopyFuncs returns the generated funcs, since we aren't registering them.
func GetGeneratedDeepCopyFuncs() []conversion.GeneratedDeepCopyFunc {
	return []conversion.GeneratedDeepCopyFunc{
		{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Simple).DeepCopyInto(out.(*Simple))
			return nil
		}, InType: reflect.TypeOf(&Simple{})},
		{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SimpleGetOptions).DeepCopyInto(out.(*SimpleGetOptions))
			return nil
		}, InType: reflect.TypeOf(&SimpleGetOptions{})},
		{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SimpleList).DeepCopyInto(out.(*SimpleList))
			return nil
		}, InType: reflect.TypeOf(&SimpleList{})},
		{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SimpleRoot).DeepCopyInto(out.(*SimpleRoot))
			return nil
		}, InType: reflect.TypeOf(&SimpleRoot{})},
		{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SimpleXGSubresource).DeepCopyInto(out.(*SimpleXGSubresource))
			return nil
		}, InType: reflect.TypeOf(&SimpleXGSubresource{})},
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Simple) DeepCopyInto(out *Simple) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new Simple.
func (x *Simple) DeepCopy() *Simple {
	if x == nil {
		return nil
	}
	out := new(Simple)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *Simple) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleGetOptions) DeepCopyInto(out *SimpleGetOptions) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SimpleGetOptions.
func (x *SimpleGetOptions) DeepCopy() *SimpleGetOptions {
	if x == nil {
		return nil
	}
	out := new(SimpleGetOptions)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *SimpleGetOptions) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleList) DeepCopyInto(out *SimpleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Simple, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SimpleList.
func (x *SimpleList) DeepCopy() *SimpleList {
	if x == nil {
		return nil
	}
	out := new(SimpleList)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *SimpleList) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleRoot) DeepCopyInto(out *SimpleRoot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SimpleRoot.
func (x *SimpleRoot) DeepCopy() *SimpleRoot {
	if x == nil {
		return nil
	}
	out := new(SimpleRoot)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *SimpleRoot) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleXGSubresource) DeepCopyInto(out *SimpleXGSubresource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SimpleXGSubresource.
func (x *SimpleXGSubresource) DeepCopy() *SimpleXGSubresource {
	if x == nil {
		return nil
	}
	out := new(SimpleXGSubresource)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *SimpleXGSubresource) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}
