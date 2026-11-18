# ai-os-aks profile

Defaults align with the AI OS Azure BOQ (private AKS, overlay networking, data services, private endpoints).

Mid-range node counts are used where the BOQ gives ranges (e.g. 3 nodes for "3–4"). Override via customer YAML → `kubernetes_cluster.node_pools` (replace merge).