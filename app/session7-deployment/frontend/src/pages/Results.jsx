import { useState, useEffect, useRef } from "react";
import {
  FileText,
  Globe,
  Server,
  Search,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { assetsAPI, resultsAPI } from "../services/api";

function Results() {
  const [assets, setAssets] = useState([]);
  const [selectedAsset, setSelectedAsset] = useState("");
  const [resultType, setResultType] = useState("all");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Separate state for each result type
  const [whoisData, setWhoisData] = useState(null);
  const [dnsData, setDnsData] = useState([]);
  const [subdomainData, setSubdomainData] = useState([]);

  // Pagination state for each type
  const [whoisPage, setWhoisPage] = useState(1);
  const [dnsPage, setDnsPage] = useState(1);
  const [subdomainPage, setSubdomainPage] = useState(1);

  // Filter/Search state for each type
  const [dnsSearch, setDnsSearch] = useState("");
  const [dnsTypeFilter, setDnsTypeFilter] = useState("");
  const [subdomainSearch, setSubdomainSearch] = useState("");
  const [subdomainActiveFilter, setSubdomainActiveFilter] = useState("");

  // Pagination metadata
  const [dnsPagination, setDnsPagination] = useState({});
  const [subdomainPagination, setSubdomainPagination] = useState({});

  const pageSize = 10;

  // Use ref to track if we're changing assets to prevent filter useEffect conflicts
  const isChangingAsset = useRef(false);

  useEffect(() => {
    loadAssets();
  }, []);

  useEffect(() => {
    if (selectedAsset) {
      // Flag that we're changing assets
      isChangingAsset.current = true;

      // Reset pagination when asset changes
      setWhoisPage(1);
      setDnsPage(1);
      setSubdomainPage(1);
      setDnsSearch("");
      setDnsTypeFilter("");
      setSubdomainSearch("");
      setSubdomainActiveFilter("");

      if (resultType === "all") {
        // Pass the asset ID explicitly to avoid stale closure issues
        loadAllResults(selectedAsset).finally(() => {
          // Reset flag after loading completes
          setTimeout(() => {
            isChangingAsset.current = false;
          }, 100);
        });
      } else {
        loadSingleTypeResults().finally(() => {
          setTimeout(() => {
            isChangingAsset.current = false;
          }, 100);
        });
      }
    }
  }, [selectedAsset, resultType]);

  // Handle filter/pagination changes in "all" mode
  useEffect(() => {
    // Don't trigger during asset change
    if (isChangingAsset.current) return;
    if (!selectedAsset || resultType !== "all") return;

    loadDNSResults();
  }, [dnsPage, dnsSearch, dnsTypeFilter]);

  useEffect(() => {
    // Don't trigger during asset change
    if (isChangingAsset.current) return;
    if (!selectedAsset || resultType !== "all") return;

    loadSubdomainResults();
  }, [subdomainPage, subdomainSearch, subdomainActiveFilter]);

  // Reload when switching to individual result types
  useEffect(() => {
    if (selectedAsset && resultType === "dns") {
      loadDNSResults();
    }
  }, [selectedAsset, resultType, dnsPage, dnsSearch, dnsTypeFilter]);

  useEffect(() => {
    if (selectedAsset && resultType === "subdomains") {
      loadSubdomainResults();
    }
  }, [
    selectedAsset,
    resultType,
    subdomainPage,
    subdomainSearch,
    subdomainActiveFilter,
  ]);

  useEffect(() => {
    if (selectedAsset && resultType === "whois") {
      loadWhoisResults();
    }
  }, [selectedAsset, resultType]);

  const loadAssets = async () => {
    try {
      const data = await assetsAPI.list({ page_size: 100 });
      setAssets(data.data || []);
      if (data.data && data.data.length > 0) {
        setSelectedAsset(data.data[0].id);
      }
    } catch (err) {
      setError(err.message);
    }
  };

  const loadAllResults = async (assetId = selectedAsset) => {
    setLoading(true);
    setError("");

    try {
      // Load WHOIS first
      await loadWhoisResults(assetId);
      // Then load DNS with reset filters
      await loadDNSResults(assetId, { page: 1, search: "", typeFilter: "" });
      // Finally load Subdomains with reset filters
      await loadSubdomainResults(assetId, {
        page: 1,
        search: "",
        activeFilter: "",
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadWhoisResults = async (assetId = selectedAsset) => {
    try {
      const data = await resultsAPI.getWHOIS(assetId);
      setWhoisData(data.data || data);
    } catch (err) {
      console.error("Failed to load WHOIS:", err);
      setWhoisData(null);
    }
  };

  const loadDNSResults = async (assetId = selectedAsset, options = {}) => {
    try {
      const {
        page = dnsPage,
        search = dnsSearch,
        typeFilter = dnsTypeFilter,
      } = options;

      const params = {
        page: page,
        page_size: pageSize,
      };
      if (search) params.search = search;
      if (typeFilter) params.type = typeFilter;

      const data = await resultsAPI.getDNS(assetId, params);
      setDnsData(data.data || []);
      // Extract pagination from response root level
      setDnsPagination({
        total: data.total || 0,
        page: data.page || page,
        page_size: data.page_size || pageSize,
        total_pages: data.total_pages || 0,
      });
    } catch (err) {
      console.error("Failed to load DNS:", err);
      // Don't clear data on error to prevent blank page
      setDnsData([]);
      setDnsPagination({});
    }
  };

  const loadSubdomainResults = async (
    assetId = selectedAsset,
    options = {},
  ) => {
    try {
      const {
        page = subdomainPage,
        search = subdomainSearch,
        activeFilter = subdomainActiveFilter,
      } = options;

      const params = {
        page: page,
        page_size: pageSize,
      };
      if (search) params.search = search;
      if (activeFilter !== "") params.active = activeFilter;

      const data = await resultsAPI.getSubdomains(assetId, params);
      setSubdomainData(data.data || []);
      // Extract pagination from response root level
      setSubdomainPagination({
        total: data.total || 0,
        page: data.page || page,
        page_size: data.page_size || pageSize,
        total_pages: data.total_pages || 0,
      });
    } catch (err) {
      console.error("Failed to load Subdomains:", err);
      // Don't clear data on error to prevent blank page
      setSubdomainData([]);
      setSubdomainPagination({});
    }
  };

  const loadSingleTypeResults = async () => {
    if (!selectedAsset) return;

    try {
      setLoading(true);
      setError("");

      if (resultType === "dns") {
        await loadDNSResults();
      } else if (resultType === "subdomains") {
        await loadSubdomainResults();
      } else if (resultType === "whois") {
        await loadWhoisResults();
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const selectedAssetData = assets.find((a) => a.id === selectedAsset);

  const renderPagination = (pagination, currentPage, setPage) => {
    if (!pagination || !pagination.total_pages || pagination.total_pages <= 1) {
      return null;
    }

    // Generate page numbers to show (max 7: first, ..., current-1, current, current+1, ..., last)
    const generatePageNumbers = () => {
      const pages = [];
      const totalPages = pagination.total_pages;

      if (totalPages <= 7) {
        // Show all pages if 7 or less
        for (let i = 1; i <= totalPages; i++) {
          pages.push(i);
        }
      } else {
        // Always show first page
        pages.push(1);

        if (currentPage > 3) {
          pages.push("...");
        }

        // Show pages around current
        for (
          let i = Math.max(2, currentPage - 1);
          i <= Math.min(currentPage + 1, totalPages - 1);
          i++
        ) {
          pages.push(i);
        }

        if (currentPage < totalPages - 2) {
          pages.push("...");
        }

        // Always show last page
        pages.push(totalPages);
      }

      return pages;
    };

    return (
      <div className="flex items-center justify-between mt-4 px-4 pb-4">
        <div className="text-sm text-muted">
          Showing {(currentPage - 1) * pageSize + 1} to{" "}
          {Math.min(currentPage * pageSize, pagination.total || 0)} of{" "}
          {pagination.total || 0} results
        </div>
        <div className="flex items-center gap-2">
          <button
            className="btn btn-secondary btn-sm"
            disabled={currentPage === 1}
            onClick={() => setPage(currentPage - 1)}
          >
            <ChevronLeft size={16} />
            Previous
          </button>

          {generatePageNumbers().map((pageNum, idx) =>
            pageNum === "..." ? (
              <span key={`ellipsis-${idx}`} className="px-2 text-muted">
                ...
              </span>
            ) : (
              <button
                key={pageNum}
                className={`btn btn-sm ${
                  pageNum === currentPage ? "btn-primary" : "btn-secondary"
                }`}
                onClick={() => setPage(pageNum)}
              >
                {pageNum}
              </button>
            ),
          )}

          <button
            className="btn btn-secondary btn-sm"
            disabled={currentPage === pagination.total_pages}
            onClick={() => setPage(currentPage + 1)}
          >
            Next
            <ChevronRight size={16} />
          </button>
        </div>
      </div>
    );
  };

  const renderDNSRecords = (records, showFilters = false) => {
    return (
      <>
        {showFilters && (
          <div className="p-4 border-b">
            <div className="grid grid-2 gap-4">
              <div>
                <label className="form-label">Search DNS Records</label>
                <div style={{ position: "relative" }}>
                  <Search
                    style={{
                      position: "absolute",
                      left: "12px",
                      top: "50%",
                      transform: "translateY(-50%)",
                      color: "var(--color-text-muted)",
                      pointerEvents: "none",
                      zIndex: 1,
                    }}
                    size={18}
                  />
                  <input
                    type="text"
                    className="form-input"
                    style={{ paddingLeft: "40px" }}
                    placeholder="Search by name or value..."
                    value={dnsSearch}
                    onChange={(e) => {
                      setDnsSearch(e.target.value);
                      setDnsPage(1);
                    }}
                  />
                </div>
              </div>
              <div>
                <label className="form-label">Filter by Type</label>
                <select
                  className="form-select"
                  value={dnsTypeFilter}
                  onChange={(e) => {
                    setDnsTypeFilter(e.target.value);
                    setDnsPage(1);
                  }}
                >
                  <option value="">All Types</option>
                  <option value="A">A Records</option>
                  <option value="AAAA">AAAA Records</option>
                  <option value="MX">MX Records</option>
                  <option value="NS">NS Records</option>
                  <option value="TXT">TXT Records</option>
                  <option value="CNAME">CNAME Records</option>
                </select>
              </div>
            </div>
          </div>
        )}

        {!records || records.length === 0 ? (
          <div className="p-4">
            <p className="text-muted">No DNS records found</p>
          </div>
        ) : (
          <>
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>Type</th>
                    <th>Name</th>
                    <th>Value</th>
                    <th>TTL</th>
                  </tr>
                </thead>
                <tbody>
                  {records.map((record, idx) => (
                    <tr key={record.id || idx}>
                      <td>
                        <span className="badge badge-primary">
                          {record.record_type}
                        </span>
                      </td>
                      <td className="font-medium">{record.name}</td>
                      <td className="text-sm">{record.value}</td>
                      <td className="text-sm text-muted">{record.ttl}s</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {showFilters &&
              renderPagination(dnsPagination, dnsPage, setDnsPage)}
          </>
        )}
      </>
    );
  };

  const renderSubdomains = (subdomains, showFilters = false) => {
    return (
      <>
        {showFilters && (
          <div className="p-4 border-b">
            <div className="grid grid-2 gap-4">
              <div>
                <label className="form-label">Search Subdomains</label>
                <div style={{ position: "relative" }}>
                  <Search
                    style={{
                      position: "absolute",
                      left: "12px",
                      top: "50%",
                      transform: "translateY(-50%)",
                      color: "var(--color-text-muted)",
                      pointerEvents: "none",
                      zIndex: 1,
                    }}
                    size={18}
                  />
                  <input
                    type="text"
                    className="form-input"
                    style={{ paddingLeft: "40px" }}
                    placeholder="Search by subdomain name..."
                    value={subdomainSearch}
                    onChange={(e) => {
                      setSubdomainSearch(e.target.value);
                      setSubdomainPage(1);
                    }}
                  />
                </div>
              </div>
              <div>
                <label className="form-label">Filter by Status</label>
                <select
                  className="form-select"
                  value={subdomainActiveFilter}
                  onChange={(e) => {
                    setSubdomainActiveFilter(e.target.value);
                    setSubdomainPage(1);
                  }}
                >
                  <option value="">All Status</option>
                  <option value="true">Active Only</option>
                  <option value="false">Inactive Only</option>
                </select>
              </div>
            </div>
          </div>
        )}

        {!subdomains || subdomains.length === 0 ? (
          <div className="p-4">
            <p className="text-muted">No subdomains found</p>
          </div>
        ) : (
          <>
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>Subdomain</th>
                    <th>Source</th>
                    <th>Active</th>
                    <th>Discovered</th>
                  </tr>
                </thead>
                <tbody>
                  {subdomains.map((subdomain, idx) => (
                    <tr key={subdomain.id || idx}>
                      <td className="font-medium">{subdomain.name}</td>
                      <td>
                        <span className="badge badge-info">
                          {subdomain.source}
                        </span>
                      </td>
                      <td>
                        <span
                          className={`badge ${
                            subdomain.is_active
                              ? "badge-success"
                              : "badge-secondary"
                          }`}
                        >
                          {subdomain.is_active ? "Yes" : "No"}
                        </span>
                      </td>
                      <td className="text-sm text-muted">
                        {subdomain.created_at
                          ? new Date(subdomain.created_at).toLocaleDateString()
                          : "N/A"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {showFilters &&
              renderPagination(
                subdomainPagination,
                subdomainPage,
                setSubdomainPage,
              )}
          </>
        )}
      </>
    );
  };

  const renderWHOIS = (whois) => {
    if (!whois) {
      return (
        <div className="p-4">
          <p className="text-muted">No WHOIS data found</p>
        </div>
      );
    }

    return (
      <div className="p-4 space-y-4">
        <div className="grid grid-2">
          <div>
            <label className="text-sm font-semibold text-muted">
              Registrar
            </label>
            <p>{whois.registrar || "N/A"}</p>
          </div>
          <div>
            <label className="text-sm font-semibold text-muted">Status</label>
            <p>{whois.status || "N/A"}</p>
          </div>
          <div>
            <label className="text-sm font-semibold text-muted">
              Created Date
            </label>
            <p>
              {whois.created_date
                ? new Date(whois.created_date).toLocaleDateString()
                : "N/A"}
            </p>
          </div>
          <div>
            <label className="text-sm font-semibold text-muted">
              Expiry Date
            </label>
            <p>
              {whois.expiry_date
                ? new Date(whois.expiry_date).toLocaleDateString()
                : "N/A"}
            </p>
          </div>
        </div>

        {whois.name_servers && (
          <div>
            <label className="text-sm font-semibold text-muted">
              Name Servers
            </label>
            <div className="flex flex-wrap gap-2 mt-2">
              {(() => {
                try {
                  const servers =
                    typeof whois.name_servers === "string"
                      ? JSON.parse(whois.name_servers)
                      : whois.name_servers;
                  return Array.isArray(servers)
                    ? servers.map((ns, idx) => (
                        <span key={idx} className="badge badge-info">
                          {ns}
                        </span>
                      ))
                    : null;
                } catch (e) {
                  console.error("Failed to parse name_servers:", e);
                  return null;
                }
              })()}
            </div>
          </div>
        )}

        {whois.raw_data && (
          <div>
            <label className="text-sm font-semibold text-muted">
              Raw WHOIS Data
            </label>
            <pre className="mt-2 p-4 bg-gray-50 rounded-md text-xs overflow-x-auto">
              {whois.raw_data}
            </pre>
          </div>
        )}
      </div>
    );
  };

  const renderAllResults = () => {
    return (
      <div className="space-y-6">
        {/* WHOIS - First */}
        <div className="card">
          <div className="card-header">
            <h3 className="card-title flex items-center">
              <Server size={20} className="mr-2" />
              WHOIS Information
            </h3>
          </div>
          {renderWHOIS(whoisData)}
        </div>

        {/* DNS Records - With pagination and filters */}
        <div className="card">
          <div className="card-header">
            <h3 className="card-title flex items-center">
              <Globe size={20} className="mr-2" />
              DNS Records ({dnsPagination.total || 0})
            </h3>
          </div>
          {renderDNSRecords(dnsData, true)}
        </div>

        {/* Subdomains - With pagination and filters */}
        <div className="card">
          <div className="card-header">
            <h3 className="card-title flex items-center">
              <Globe size={20} className="mr-2" />
              Subdomains ({subdomainPagination.total || 0})
            </h3>
          </div>
          {renderSubdomains(subdomainData, true)}
        </div>
      </div>
    );
  };

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Scan Results</h1>
        <p className="page-description">
          View and analyze reconnaissance data collected from scans
        </p>
      </div>

      {error && <div className="alert alert-error mb-4">{error}</div>}

      {/* Filters */}
      <div className="card mb-4">
        <div className="grid grid-2 gap-4">
          <div className="form-group">
            <label className="form-label">Select Asset</label>
            <select
              className="form-select"
              value={selectedAsset}
              onChange={(e) => setSelectedAsset(e.target.value)}
            >
              {assets.length === 0 ? (
                <option>No assets available</option>
              ) : (
                assets.map((asset) => (
                  <option key={asset.id} value={asset.id}>
                    {asset.name} ({asset.type})
                  </option>
                ))
              )}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Result Type</label>
            <select
              className="form-select"
              value={resultType}
              onChange={(e) => setResultType(e.target.value)}
            >
              <option value="all">All Results</option>
              <option value="dns">DNS Records</option>
              <option value="subdomains">Subdomains</option>
              <option value="whois">WHOIS Information</option>
            </select>
          </div>
        </div>

        {selectedAssetData && (
          <div className="mt-4 p-4 bg-gray-50 rounded-md">
            <h4 className="font-semibold text-sm mb-2">Asset Details:</h4>
            <div className="flex items-center gap-4 text-sm text-muted">
              <span>
                <strong>Name:</strong> {selectedAssetData.name}
              </span>
              <span>
                <strong>Type:</strong> {selectedAssetData.type}
              </span>
              <span>
                <strong>Status:</strong> {selectedAssetData.status}
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Results Display */}
      {loading ? (
        <div className="card">
          <div className="loading">
            <div className="spinner"></div>
            <span>Loading results...</span>
          </div>
        </div>
      ) : !selectedAsset ? (
        <div className="card">
          <div className="empty-state">
            <FileText className="empty-state-icon" size={64} />
            <h3 className="empty-state-title">No asset selected</h3>
            <p className="empty-state-description">
              Select an asset to view its scan results
            </p>
          </div>
        </div>
      ) : (
        <>
          {resultType === "all" && renderAllResults()}

          {resultType === "dns" && (
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">
                  <Globe size={20} className="inline mr-2" />
                  DNS Records
                </h3>
              </div>
              {renderDNSRecords(dnsData, true)}
            </div>
          )}

          {resultType === "subdomains" && (
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">
                  <Globe size={20} className="inline mr-2" />
                  Subdomains
                </h3>
              </div>
              {renderSubdomains(subdomainData, true)}
            </div>
          )}

          {resultType === "whois" && (
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">
                  <Server size={20} className="inline mr-2" />
                  WHOIS Information
                </h3>
              </div>
              {renderWHOIS(whoisData)}
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default Results;
