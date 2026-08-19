# Security Policy

## Security model

Halko controls a single wood drying kiln from a dedicated machine — in practice
a Raspberry Pi — on a private, trusted network. It is an appliance for one
operator and one kiln, not a multi-user service, and it is built throughout on
that assumption.

**Halko has no authentication, no authorization, and no transport security.**
That is a deliberate choice for the environment it was built for, not an
oversight. Anyone who can reach a Halko service over the network can do
anything that service can do.

Specifically:

- The REST APIs are plain HTTP. There is no TLS anywhere in the system.
- No endpoint requires credentials of any kind.
- Every service sends `Access-Control-Allow-Origin: *`, so any page open in a
  browser on the network can call the APIs.
- Services listen on all interfaces, not only loopback.
- `dbusunit` runs as root and exposes unauthenticated endpoints that shut down
  or reboot the host and start or stop VPN connections. The other services run
  unprivileged as the `halko` user.
- Stored programs, run history and execution logs are written to the
  filesystem from data supplied over the API.

## Deploying it safely

The network is the entire security boundary. These are requirements rather
than recommendations:

1. **Never expose Halko to the internet.** No port forwarding, no reverse
   proxy on a public address, no cloud tunnel.
2. **Put it on a segregated network** — its own VLAN or subnet, reachable only
   from the machines that operate the kiln.
3. **Reach it remotely over a VPN**, never by opening ports. The VPN control
   in `dbusunit` exists for exactly this.
4. **Treat network access as full control of the machine.** Anyone who can
   reach `dbusunit` can power off the host mid-run.
5. **Keep the Shelly on its dedicated link**, as described in
   [RASPBERRY_PI.md](RASPBERRY_PI.md), rather than on the general network.

If you need to run Halko somewhere untrusted, it is not the right software as
it stands. You would have to put an authenticating proxy in front of every
service and terminate TLS there — and no part of Halko has been designed or
reviewed for that.

## Supported versions

Only the most recent release is supported. Halko is versioned as described in
[CONTRIBUTING.md](CONTRIBUTING.md); fixes go into a new version rather than
being backported.

| Version | Supported |
| ------- | --------- |
| 1.0.x   | Yes       |
| < 1.0   | No        |

## What is not a vulnerability

Everything in the Security model section above is documented and intended.
Reports amounting to "the API has no authentication", "traffic is not
encrypted", "CORS allows any origin", or "the service listens on all
interfaces" describe the design, and will be closed as such.

What is worth reporting is anything that lets someone already on the trusted
network do something the design does not intend, or that undermines the
isolation the deployment guidance depends on. For example:

- escaping `base_path` when writing programs, status files or logs
- getting a service to execute arbitrary commands on the host
- crashing or wedging a service with input a kiln operator could plausibly
  send, in a way that abandons a running program
- a dependency carrying a known, reachable vulnerability

## Reporting a vulnerability

Use GitHub's private vulnerability reporting: the repository's **Security**
tab, then **Report a vulnerability**. The report stays private between you and
the maintainer until an advisory is published.

Please do not open a public issue for a security problem in the first
instance.

Include the version you are running, how to reproduce the problem, and what an
attacker on the network could achieve with it.

Halko is a small project maintained by one person alongside other work. Expect
an acknowledgement in a couple of weeks rather than a couple of days, and no
bounty — there is no budget for one.
