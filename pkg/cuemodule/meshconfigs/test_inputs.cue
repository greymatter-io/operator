package meshconfigs

// This file is for manually testing various inputs with the cue CLI.
// i.e. "cue eval -e service.routes"

// mesh: {
//   metadata: {
//     name: "mymesh"
//   }
//   spec: {
//     install_namespace: "greymatter"
//     release_version: "1.7"
//     zone: "myzone"
//   }
// }

// workload: {
//   metadata: {
//     name: "example"
//     namespace: "myns"
//     annotations: {
//       "greymatter.io/egress-http-local": """
//       ["cluster1","cluster2"]
//       """
//       "greymatter.io/egress-http-external": """
//       [
//         {
//           "name": "cluster3",
//           "host": "cluster3.org",
//           "port": 80
//         },
//         {
//           "name": "cluster4",
//           "host": "cluster4.org",
//           "port": 80
//         }
//       ]
//       """
//       "greymatter.io/egress-tcp-local": """
//       ["cluster5","cluster6"]
//       """
//       "greymatter.io/egress-tcp-external": """
//       [
//         {
//           "name": "cluster7",
//           "host": "cluster7.org",
//           "port": 80
//         },
//         {
//           "name": "cluster8",
//           "host": "cluster8.org",
//           "port": 80
//         }
//       ]
//       """
//     }
//   }
//   spec: {
//     template: {
//       spec: {
//         containers: [
//           {
//             name: "c1"
//             ports: [
//               {name: "server1", containerPort: 5555},
//               {name: "server2", containerPort: 8080},
//             ]
//           },
//          {
//            name: "c2"
//            ports: [
//              {name: "server3", containerPort: 3000},
//            ]
//          },
//          {
//            name: "unnamed"
//            ports: [
//              {containerPort: 8081},
//            ]
//          },
//         ]
//       }
//     }
//   }
// }
