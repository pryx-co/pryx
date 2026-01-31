// MCP Service for HTTP API calls
export interface McpServer {
  id: string;
  name: string;
  url: string;
  transport: string;
  status: "connected" | "error" | "disabled" | "connecting";
  enabled: boolean;
  securityRating: "A" | "B" | "C" | "D" | "F";
  tools: McpTool[];
  description?: string;
  category?: string;
  author?: string;
  version?: string;
}

export interface McpTool {
  name: string;
  description: string;
  parameters?: Record<string, any>;
}

export interface CuratedServer {
  id: string;
  name: string;
  description: string;
  author: string;
  version: string;
  category: string;
  transport: string;
  url: string;
  tools: McpTool[];
  securityRating: "A" | "B" | "C" | "D" | "F";
  requirements?: string[];
}

export interface ValidationResult {
  valid: boolean;
  url: string;
  securityRating: "A" | "B" | "C" | "D" | "F";
  warnings: string[];
  errors: string[];
}

const API_BASE = "http://localhost:3000";

export class McpService {
  async getServers(): Promise<McpServer[]> {
    try {
      const response = await fetch(`${API_BASE}/mcp/discovery/custom`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return await response.json();
    } catch (error) {
      console.error("Failed to fetch MCP servers:", error);
      return [];
    }
  }

  async getCuratedServers(): Promise<CuratedServer[]> {
    try {
      const response = await fetch(`${API_BASE}/mcp/discovery/curated`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return await response.json();
    } catch (error) {
      console.error("Failed to fetch curated servers:", error);
      return [];
    }
  }

  async addServer(server: { name: string; url: string; transport: string }): Promise<McpServer> {
    const response = await fetch(`${API_BASE}/mcp/discovery/custom`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(server),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Failed to add server");
    }

    return await response.json();
  }

  async addCuratedServer(serverId: string): Promise<McpServer> {
    const response = await fetch(`${API_BASE}/mcp/discovery/custom`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        curatedId: serverId,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Failed to add curated server");
    }

    return await response.json();
  }

  async deleteServer(serverId: string): Promise<void> {
    const response = await fetch(`${API_BASE}/mcp/discovery/custom/${serverId}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Failed to delete server");
    }
  }

  async toggleServer(serverId: string, enabled: boolean): Promise<void> {
    const response = await fetch(`${API_BASE}/mcp/discovery/custom/${serverId}/toggle`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ enabled }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Failed to toggle server");
    }
  }

  async validateUrl(url: string, transport: string): Promise<ValidationResult> {
    const response = await fetch(`${API_BASE}/mcp/discovery/validate`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url, transport }),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Validation failed");
    }

    return await response.json();
  }

  async testTool(serverId: string, toolName: string, parameters: Record<string, any>): Promise<any> {
    const response = await fetch(`${API_BASE}/mcp/servers/${serverId}/tools/${toolName}/test`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(parameters),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || "Tool test failed");
    }

    return await response.json();
  }
}
