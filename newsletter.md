# Fedora CoreOS Newsletter: October 9 - October 16, 2026

Welcome to the Fedora CoreOS weekly newsletter! We're pleased to share the latest activities and contributions from our community. Your dedication helps shape Fedora CoreOS, and we're grateful for everyone's efforts.

This past week, the project saw significant activity, with **31 pull requests merged** across our repositories and **1 community meeting** held.

---

## Highlights

### Notable Contributions

We'd like to highlight some of the impactful contributions from the past week:

*   **ci: re-enable `go fix` in go pipeline** ([PR #2346](https://github.com/coreos/ignition/pull/2346)) by @cadejacobson
    *   This change re-enables the `go fix` tool in the continuous integration pipeline for Ignition, ensuring code style and practices are consistently applied.
*   **kola: add KubeVirt platform support** ([PR #4507](https://github.com/coreos/coreos-assembler/pull/4507)) by @qinqon
    *   Support for the KubeVirt platform has been added to Kola, enhancing our testing capabilities for Fedora CoreOS within Kubernetes environments.
*   **internal/partitioners: add sfdisk partitioning backend** ([PR #2337](https://github.com/coreos/ignition/pull/2337)) by @prestist
    *   Ignition now includes `sfdisk` as an additional partitioning backend, offering more flexibility in disk management during provisioning.
*   **butane: Add THREAT_MODEL.md** ([PR #2340](https://github.com/coreos/ignition/pull/2340)) by @PeaceRebel
    *   A threat model document has been added to Butane, outlining potential security risks and considerations for the project.
*   **azure: add unittests for OVF logic** ([PR #2345](https://github.com/coreos/ignition/pull/2345)) by @cadejacobson
    *   New unit tests have been introduced for the OVF (Open Virtualization Format) logic within our Azure-related code, improving reliability.

### Thanks to Our Contributors

A big thank you to all the individuals who contributed to Fedora CoreOS this week! Your work is invaluable:

*   @Ar-maan05
*   @PeaceRebel
*   @RiRa12621
*   @Rolv-Apneseth
*   @alexlarsson
*   @bwqr
*   @cadejacobson
*   @cverna
*   @jbtrystram
*   @nikita-dubrovskii
*   @prestist
*   @qinqon
*   @ravanelli
*   @w453y

---

## Community Meetings

**Meeting on 2026-10-16**
The community held its weekly meeting on October 16th. You can review the full discussion and topics covered in the [meeting log](https://meetbot.fedoraproject.org).

---

## Get Involved

We welcome everyone to participate in the Fedora CoreOS project! Whether you're a developer, tester, documenter, or user, there are many ways to contribute.

*   **Contribute Code:** Find issues and contribute on our [GitHub repositories](https://github.com/coreos).
*   **Chat with Us:** Join our community on [Matrix](https://chat.fedoraproject.org/#/room/#coreos:fedoraproject.org) or [IRC](irc://irc.libera.chat/#fedora-coreos).
*   **Mailing List:** Subscribe to our [mailing lists](https://lists.fedoraproject.org/archives/list/coreos@lists.fedoraproject.org/) for announcements and discussions.
*   **Documentation:** Help improve our [documentation](https://docs.fedoraproject.org/en-US/fedora-coreos/).

Thank you for being a part of the Fedora CoreOS community!