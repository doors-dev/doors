// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

export const id: string = document.currentScript!.id
export const prefix: string = document.currentScript!.dataset.prefix!
export const rootId: number = Number(document.currentScript!.dataset.root)
export const ttl: number = Number(document.currentScript!.dataset.ttl)
export const disconnectAfter: number = Number(document.currentScript!.dataset.disconnect)
export const requestTimeout: number = Number(document.currentScript!.dataset.request)
export const solitairePing: number = Number(document.currentScript!.dataset.ping)
