// Software types

export interface Software {
  id: number;
  cpe_string: string;
  vendor: string;
  product_name: string;
  version?: string;
  edition?: string;
  target_hw?: string;
  lang?: string;
  title_formatted: string;
  created_at: string;
  updated_at: string;
}

export interface AssetSoftware {
  id: number;
  tenant_id: number;
  asset_id: number;
  software_id: number;
  source: string;
  install_path?: string;
  first_seen_at: string;
  last_seen_at: string;
  created_at: string;
  updated_at: string;
  // Joined fields from software table
  cpe_string: string;
  vendor: string;
  product_name: string;
  version?: string;
  edition?: string;
  title_formatted: string;
}

export interface SoftwareSummary {
  software_id: number;
  cpe_string: string;
  vendor: string;
  product_name: string;
  version?: string;
  title_formatted: string;
  install_count: number;
}

export interface SoftwareDetails {
  software: Software;
  affected_assets: Array<{
    asset_id: number;
    canonical_name: string;
    is_active: boolean;
    first_seen_at: string;
    last_seen_at: string;
    install_path?: string;
  }>;
  total_assets: number;
  affected_findings: {
    total_findings: number;
    critical_count: number;
    high_count: number;
    medium_count: number;
    low_count: number;
  };
}

export interface SoftwareListResponse {
  software: SoftwareSummary[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
  };
}

export interface SoftwareFilters {
  vendor?: string;
  product?: string;
  version?: string;
  cpe?: string;
}
