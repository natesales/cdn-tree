# Controller Bootstrap

The cbootstrap process is used to bring up new controller nodes and adding them to the network. Control plane nodes are managed using the Ansible playbooks contained in this directory.

### Preflight
1. Configure IPv4 and IPv6 addresses on the controller
2. Add the controller SSH key to the root user
3. Set PermitRootLogin to prohibit-password in sshd_config
4. Run install playbook

### Initialize a replica set
`rs.initiate({_id:"packetframe", members: [{_id: 1, host: "172.16.17.1:27017"}, {_id: 2, host: "172.16.17.2:27017"}, {_id: 3, host: "172.16.17.3:27017"}]});`
