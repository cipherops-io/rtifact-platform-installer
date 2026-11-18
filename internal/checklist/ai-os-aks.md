Manual steps not covered by Azure RM automation:

- Raise subscription quotas (Dsv5, Esv5, Fsv2, public IPs, managed disks).
- MongoDB Atlas M30 project, peering or Private Link to VNet.
- Verify Communication Services email domain (SPF, DKIM, DMARC).
- Microsoft Foundry model deployments and token quotas.
- Configure Azure Firewall or proxy with Annex A outbound allow-list.
- DNS for apps and ops domains; TLS (cert-manager DNS-01 or Key Vault certs).
- VPN / Bastion access for operations team.
- Data source connectivity and credentials in Key Vault.
- Backup `tofu.tfstate` if using local backend.
- Approve Claude inbound agent paths and outbound Anthropic endpoints if required.