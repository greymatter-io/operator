package base

// This file is for manually testing various inputs with the cue CLI.
// i.e. "cue eval -e mesh.spec"

// mesh: {
//   metadata: {
//     name: "mymesh"
//   }
//   spec: {
//     install_namespace: "greymatter"
//     release_version: "1.7"
//     zone: "myzone"
//     image_pull_secrets: [
//         "test1",
//         "test2"
//     ]
//     images: {
//         catalog: "docker.greymatter.io/development/gm-catalog:3.0.0"
//     }
//   }
// }
