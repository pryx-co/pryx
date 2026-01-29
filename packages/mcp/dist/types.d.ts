import { z } from 'zod';
export declare const TransportType: z.ZodEnum<["stdio", "sse", "websocket"]>;
export declare const ServerSource: z.ZodEnum<["manual", "curated", "marketplace"]>;
export declare const ToolDefinitionSchema: z.ZodObject<{
    name: z.ZodString;
    description: z.ZodString;
    inputSchema: z.ZodRecord<z.ZodString, z.ZodUnknown>;
}, "strip", z.ZodTypeAny, {
    name: string;
    description: string;
    inputSchema: Record<string, unknown>;
}, {
    name: string;
    description: string;
    inputSchema: Record<string, unknown>;
}>;
export declare const ResourceDefinitionSchema: z.ZodObject<{
    uri: z.ZodString;
    name: z.ZodString;
    mimeType: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    name: string;
    uri: string;
    mimeType?: string | undefined;
}, {
    name: string;
    uri: string;
    mimeType?: string | undefined;
}>;
export declare const ArgumentDefinitionSchema: z.ZodObject<{
    name: z.ZodString;
    description: z.ZodString;
    required: z.ZodDefault<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    name: string;
    description: string;
    required: boolean;
}, {
    name: string;
    description: string;
    required?: boolean | undefined;
}>;
export declare const PromptDefinitionSchema: z.ZodObject<{
    name: z.ZodString;
    description: z.ZodString;
    arguments: z.ZodOptional<z.ZodArray<z.ZodObject<{
        name: z.ZodString;
        description: z.ZodString;
        required: z.ZodDefault<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        name: string;
        description: string;
        required: boolean;
    }, {
        name: string;
        description: string;
        required?: boolean | undefined;
    }>, "many">>;
}, "strip", z.ZodTypeAny, {
    name: string;
    description: string;
    arguments?: {
        name: string;
        description: string;
        required: boolean;
    }[] | undefined;
}, {
    name: string;
    description: string;
    arguments?: {
        name: string;
        description: string;
        required?: boolean | undefined;
    }[] | undefined;
}>;
export declare const CapabilitiesSchema: z.ZodObject<{
    tools: z.ZodDefault<z.ZodArray<z.ZodObject<{
        name: z.ZodString;
        description: z.ZodString;
        inputSchema: z.ZodRecord<z.ZodString, z.ZodUnknown>;
    }, "strip", z.ZodTypeAny, {
        name: string;
        description: string;
        inputSchema: Record<string, unknown>;
    }, {
        name: string;
        description: string;
        inputSchema: Record<string, unknown>;
    }>, "many">>;
    resources: z.ZodDefault<z.ZodArray<z.ZodObject<{
        uri: z.ZodString;
        name: z.ZodString;
        mimeType: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        name: string;
        uri: string;
        mimeType?: string | undefined;
    }, {
        name: string;
        uri: string;
        mimeType?: string | undefined;
    }>, "many">>;
    prompts: z.ZodDefault<z.ZodArray<z.ZodObject<{
        name: z.ZodString;
        description: z.ZodString;
        arguments: z.ZodOptional<z.ZodArray<z.ZodObject<{
            name: z.ZodString;
            description: z.ZodString;
            required: z.ZodDefault<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            description: string;
            required: boolean;
        }, {
            name: string;
            description: string;
            required?: boolean | undefined;
        }>, "many">>;
    }, "strip", z.ZodTypeAny, {
        name: string;
        description: string;
        arguments?: {
            name: string;
            description: string;
            required: boolean;
        }[] | undefined;
    }, {
        name: string;
        description: string;
        arguments?: {
            name: string;
            description: string;
            required?: boolean | undefined;
        }[] | undefined;
    }>, "many">>;
}, "strip", z.ZodTypeAny, {
    tools: {
        name: string;
        description: string;
        inputSchema: Record<string, unknown>;
    }[];
    resources: {
        name: string;
        uri: string;
        mimeType?: string | undefined;
    }[];
    prompts: {
        name: string;
        description: string;
        arguments?: {
            name: string;
            description: string;
            required: boolean;
        }[] | undefined;
    }[];
}, {
    tools?: {
        name: string;
        description: string;
        inputSchema: Record<string, unknown>;
    }[] | undefined;
    resources?: {
        name: string;
        uri: string;
        mimeType?: string | undefined;
    }[] | undefined;
    prompts?: {
        name: string;
        description: string;
        arguments?: {
            name: string;
            description: string;
            required?: boolean | undefined;
        }[] | undefined;
    }[] | undefined;
}>;
export declare const StdioTransportSchema: z.ZodObject<{
    type: z.ZodLiteral<"stdio">;
    command: z.ZodString;
    args: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    env: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
    cwd: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    type: "stdio";
    command: string;
    args: string[];
    env: Record<string, string>;
    cwd?: string | undefined;
}, {
    type: "stdio";
    command: string;
    args?: string[] | undefined;
    env?: Record<string, string> | undefined;
    cwd?: string | undefined;
}>;
export declare const SSETransportSchema: z.ZodObject<{
    type: z.ZodLiteral<"sse">;
    url: z.ZodString;
    headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
}, "strip", z.ZodTypeAny, {
    type: "sse";
    url: string;
    headers: Record<string, string>;
}, {
    type: "sse";
    url: string;
    headers?: Record<string, string> | undefined;
}>;
export declare const WebSocketTransportSchema: z.ZodObject<{
    type: z.ZodLiteral<"websocket">;
    url: z.ZodString;
    headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
}, "strip", z.ZodTypeAny, {
    type: "websocket";
    url: string;
    headers: Record<string, string>;
}, {
    type: "websocket";
    url: string;
    headers?: Record<string, string> | undefined;
}>;
export declare const TransportConfigSchema: z.ZodUnion<[z.ZodObject<{
    type: z.ZodLiteral<"stdio">;
    command: z.ZodString;
    args: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    env: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
    cwd: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    type: "stdio";
    command: string;
    args: string[];
    env: Record<string, string>;
    cwd?: string | undefined;
}, {
    type: "stdio";
    command: string;
    args?: string[] | undefined;
    env?: Record<string, string> | undefined;
    cwd?: string | undefined;
}>, z.ZodObject<{
    type: z.ZodLiteral<"sse">;
    url: z.ZodString;
    headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
}, "strip", z.ZodTypeAny, {
    type: "sse";
    url: string;
    headers: Record<string, string>;
}, {
    type: "sse";
    url: string;
    headers?: Record<string, string> | undefined;
}>, z.ZodObject<{
    type: z.ZodLiteral<"websocket">;
    url: z.ZodString;
    headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
}, "strip", z.ZodTypeAny, {
    type: "websocket";
    url: string;
    headers: Record<string, string>;
}, {
    type: "websocket";
    url: string;
    headers?: Record<string, string> | undefined;
}>]>;
export declare const ServerSettingsSchema: z.ZodObject<{
    autoConnect: z.ZodDefault<z.ZodBoolean>;
    timeout: z.ZodDefault<z.ZodNumber>;
    reconnect: z.ZodDefault<z.ZodBoolean>;
    maxReconnectAttempts: z.ZodDefault<z.ZodNumber>;
    fallbackServers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
}, "strip", z.ZodTypeAny, {
    autoConnect: boolean;
    timeout: number;
    reconnect: boolean;
    maxReconnectAttempts: number;
    fallbackServers: string[];
}, {
    autoConnect?: boolean | undefined;
    timeout?: number | undefined;
    reconnect?: boolean | undefined;
    maxReconnectAttempts?: number | undefined;
    fallbackServers?: string[] | undefined;
}>;
export declare const ConnectionStatusSchema: z.ZodObject<{
    connected: z.ZodBoolean;
    lastConnected: z.ZodOptional<z.ZodString>;
    lastError: z.ZodOptional<z.ZodString>;
    reconnectAttempts: z.ZodDefault<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    connected: boolean;
    reconnectAttempts: number;
    lastConnected?: string | undefined;
    lastError?: string | undefined;
}, {
    connected: boolean;
    lastConnected?: string | undefined;
    lastError?: string | undefined;
    reconnectAttempts?: number | undefined;
}>;
export declare const MCPServerConfigSchema: z.ZodObject<{
    id: z.ZodString;
    name: z.ZodString;
    enabled: z.ZodDefault<z.ZodBoolean>;
    transport: z.ZodUnion<[z.ZodObject<{
        type: z.ZodLiteral<"stdio">;
        command: z.ZodString;
        args: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        env: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
        cwd: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        type: "stdio";
        command: string;
        args: string[];
        env: Record<string, string>;
        cwd?: string | undefined;
    }, {
        type: "stdio";
        command: string;
        args?: string[] | undefined;
        env?: Record<string, string> | undefined;
        cwd?: string | undefined;
    }>, z.ZodObject<{
        type: z.ZodLiteral<"sse">;
        url: z.ZodString;
        headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        type: "sse";
        url: string;
        headers: Record<string, string>;
    }, {
        type: "sse";
        url: string;
        headers?: Record<string, string> | undefined;
    }>, z.ZodObject<{
        type: z.ZodLiteral<"websocket">;
        url: z.ZodString;
        headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        type: "websocket";
        url: string;
        headers: Record<string, string>;
    }, {
        type: "websocket";
        url: string;
        headers?: Record<string, string> | undefined;
    }>]>;
    capabilities: z.ZodOptional<z.ZodObject<{
        tools: z.ZodDefault<z.ZodArray<z.ZodObject<{
            name: z.ZodString;
            description: z.ZodString;
            inputSchema: z.ZodRecord<z.ZodString, z.ZodUnknown>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }, {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }>, "many">>;
        resources: z.ZodDefault<z.ZodArray<z.ZodObject<{
            uri: z.ZodString;
            name: z.ZodString;
            mimeType: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }, {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }>, "many">>;
        prompts: z.ZodDefault<z.ZodArray<z.ZodObject<{
            name: z.ZodString;
            description: z.ZodString;
            arguments: z.ZodOptional<z.ZodArray<z.ZodObject<{
                name: z.ZodString;
                description: z.ZodString;
                required: z.ZodDefault<z.ZodBoolean>;
            }, "strip", z.ZodTypeAny, {
                name: string;
                description: string;
                required: boolean;
            }, {
                name: string;
                description: string;
                required?: boolean | undefined;
            }>, "many">>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required: boolean;
            }[] | undefined;
        }, {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required?: boolean | undefined;
            }[] | undefined;
        }>, "many">>;
    }, "strip", z.ZodTypeAny, {
        tools: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[];
        resources: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[];
        prompts: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required: boolean;
            }[] | undefined;
        }[];
    }, {
        tools?: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[] | undefined;
        resources?: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[] | undefined;
        prompts?: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required?: boolean | undefined;
            }[] | undefined;
        }[] | undefined;
    }>>;
    source: z.ZodDefault<z.ZodEnum<["manual", "curated", "marketplace"]>>;
    settings: z.ZodDefault<z.ZodObject<{
        autoConnect: z.ZodDefault<z.ZodBoolean>;
        timeout: z.ZodDefault<z.ZodNumber>;
        reconnect: z.ZodDefault<z.ZodBoolean>;
        maxReconnectAttempts: z.ZodDefault<z.ZodNumber>;
        fallbackServers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    }, "strip", z.ZodTypeAny, {
        autoConnect: boolean;
        timeout: number;
        reconnect: boolean;
        maxReconnectAttempts: number;
        fallbackServers: string[];
    }, {
        autoConnect?: boolean | undefined;
        timeout?: number | undefined;
        reconnect?: boolean | undefined;
        maxReconnectAttempts?: number | undefined;
        fallbackServers?: string[] | undefined;
    }>>;
    status: z.ZodOptional<z.ZodObject<{
        connected: z.ZodBoolean;
        lastConnected: z.ZodOptional<z.ZodString>;
        lastError: z.ZodOptional<z.ZodString>;
        reconnectAttempts: z.ZodDefault<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        connected: boolean;
        reconnectAttempts: number;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
    }, {
        connected: boolean;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
        reconnectAttempts?: number | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    name: string;
    id: string;
    enabled: boolean;
    transport: {
        type: "stdio";
        command: string;
        args: string[];
        env: Record<string, string>;
        cwd?: string | undefined;
    } | {
        type: "sse";
        url: string;
        headers: Record<string, string>;
    } | {
        type: "websocket";
        url: string;
        headers: Record<string, string>;
    };
    source: "manual" | "curated" | "marketplace";
    settings: {
        autoConnect: boolean;
        timeout: number;
        reconnect: boolean;
        maxReconnectAttempts: number;
        fallbackServers: string[];
    };
    status?: {
        connected: boolean;
        reconnectAttempts: number;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
    } | undefined;
    capabilities?: {
        tools: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[];
        resources: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[];
        prompts: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required: boolean;
            }[] | undefined;
        }[];
    } | undefined;
}, {
    name: string;
    id: string;
    transport: {
        type: "stdio";
        command: string;
        args?: string[] | undefined;
        env?: Record<string, string> | undefined;
        cwd?: string | undefined;
    } | {
        type: "sse";
        url: string;
        headers?: Record<string, string> | undefined;
    } | {
        type: "websocket";
        url: string;
        headers?: Record<string, string> | undefined;
    };
    status?: {
        connected: boolean;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
        reconnectAttempts?: number | undefined;
    } | undefined;
    enabled?: boolean | undefined;
    capabilities?: {
        tools?: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[] | undefined;
        resources?: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[] | undefined;
        prompts?: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required?: boolean | undefined;
            }[] | undefined;
        }[] | undefined;
    } | undefined;
    source?: "manual" | "curated" | "marketplace" | undefined;
    settings?: {
        autoConnect?: boolean | undefined;
        timeout?: number | undefined;
        reconnect?: boolean | undefined;
        maxReconnectAttempts?: number | undefined;
        fallbackServers?: string[] | undefined;
    } | undefined;
}>;
export declare const MCPServersConfigSchema: z.ZodObject<{
    version: z.ZodDefault<z.ZodNumber>;
    servers: z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        name: z.ZodString;
        enabled: z.ZodDefault<z.ZodBoolean>;
        transport: z.ZodUnion<[z.ZodObject<{
            type: z.ZodLiteral<"stdio">;
            command: z.ZodString;
            args: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
            env: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
            cwd: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            type: "stdio";
            command: string;
            args: string[];
            env: Record<string, string>;
            cwd?: string | undefined;
        }, {
            type: "stdio";
            command: string;
            args?: string[] | undefined;
            env?: Record<string, string> | undefined;
            cwd?: string | undefined;
        }>, z.ZodObject<{
            type: z.ZodLiteral<"sse">;
            url: z.ZodString;
            headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
        }, "strip", z.ZodTypeAny, {
            type: "sse";
            url: string;
            headers: Record<string, string>;
        }, {
            type: "sse";
            url: string;
            headers?: Record<string, string> | undefined;
        }>, z.ZodObject<{
            type: z.ZodLiteral<"websocket">;
            url: z.ZodString;
            headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
        }, "strip", z.ZodTypeAny, {
            type: "websocket";
            url: string;
            headers: Record<string, string>;
        }, {
            type: "websocket";
            url: string;
            headers?: Record<string, string> | undefined;
        }>]>;
        capabilities: z.ZodOptional<z.ZodObject<{
            tools: z.ZodDefault<z.ZodArray<z.ZodObject<{
                name: z.ZodString;
                description: z.ZodString;
                inputSchema: z.ZodRecord<z.ZodString, z.ZodUnknown>;
            }, "strip", z.ZodTypeAny, {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }, {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }>, "many">>;
            resources: z.ZodDefault<z.ZodArray<z.ZodObject<{
                uri: z.ZodString;
                name: z.ZodString;
                mimeType: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }, {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }>, "many">>;
            prompts: z.ZodDefault<z.ZodArray<z.ZodObject<{
                name: z.ZodString;
                description: z.ZodString;
                arguments: z.ZodOptional<z.ZodArray<z.ZodObject<{
                    name: z.ZodString;
                    description: z.ZodString;
                    required: z.ZodDefault<z.ZodBoolean>;
                }, "strip", z.ZodTypeAny, {
                    name: string;
                    description: string;
                    required: boolean;
                }, {
                    name: string;
                    description: string;
                    required?: boolean | undefined;
                }>, "many">>;
            }, "strip", z.ZodTypeAny, {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required: boolean;
                }[] | undefined;
            }, {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required?: boolean | undefined;
                }[] | undefined;
            }>, "many">>;
        }, "strip", z.ZodTypeAny, {
            tools: {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }[];
            resources: {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }[];
            prompts: {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required: boolean;
                }[] | undefined;
            }[];
        }, {
            tools?: {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }[] | undefined;
            resources?: {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }[] | undefined;
            prompts?: {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required?: boolean | undefined;
                }[] | undefined;
            }[] | undefined;
        }>>;
        source: z.ZodDefault<z.ZodEnum<["manual", "curated", "marketplace"]>>;
        settings: z.ZodDefault<z.ZodObject<{
            autoConnect: z.ZodDefault<z.ZodBoolean>;
            timeout: z.ZodDefault<z.ZodNumber>;
            reconnect: z.ZodDefault<z.ZodBoolean>;
            maxReconnectAttempts: z.ZodDefault<z.ZodNumber>;
            fallbackServers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        }, "strip", z.ZodTypeAny, {
            autoConnect: boolean;
            timeout: number;
            reconnect: boolean;
            maxReconnectAttempts: number;
            fallbackServers: string[];
        }, {
            autoConnect?: boolean | undefined;
            timeout?: number | undefined;
            reconnect?: boolean | undefined;
            maxReconnectAttempts?: number | undefined;
            fallbackServers?: string[] | undefined;
        }>>;
        status: z.ZodOptional<z.ZodObject<{
            connected: z.ZodBoolean;
            lastConnected: z.ZodOptional<z.ZodString>;
            lastError: z.ZodOptional<z.ZodString>;
            reconnectAttempts: z.ZodDefault<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            connected: boolean;
            reconnectAttempts: number;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
        }, {
            connected: boolean;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
            reconnectAttempts?: number | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        name: string;
        id: string;
        enabled: boolean;
        transport: {
            type: "stdio";
            command: string;
            args: string[];
            env: Record<string, string>;
            cwd?: string | undefined;
        } | {
            type: "sse";
            url: string;
            headers: Record<string, string>;
        } | {
            type: "websocket";
            url: string;
            headers: Record<string, string>;
        };
        source: "manual" | "curated" | "marketplace";
        settings: {
            autoConnect: boolean;
            timeout: number;
            reconnect: boolean;
            maxReconnectAttempts: number;
            fallbackServers: string[];
        };
        status?: {
            connected: boolean;
            reconnectAttempts: number;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
        } | undefined;
        capabilities?: {
            tools: {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }[];
            resources: {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }[];
            prompts: {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required: boolean;
                }[] | undefined;
            }[];
        } | undefined;
    }, {
        name: string;
        id: string;
        transport: {
            type: "stdio";
            command: string;
            args?: string[] | undefined;
            env?: Record<string, string> | undefined;
            cwd?: string | undefined;
        } | {
            type: "sse";
            url: string;
            headers?: Record<string, string> | undefined;
        } | {
            type: "websocket";
            url: string;
            headers?: Record<string, string> | undefined;
        };
        status?: {
            connected: boolean;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
            reconnectAttempts?: number | undefined;
        } | undefined;
        enabled?: boolean | undefined;
        capabilities?: {
            tools?: {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }[] | undefined;
            resources?: {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }[] | undefined;
            prompts?: {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required?: boolean | undefined;
                }[] | undefined;
            }[] | undefined;
        } | undefined;
        source?: "manual" | "curated" | "marketplace" | undefined;
        settings?: {
            autoConnect?: boolean | undefined;
            timeout?: number | undefined;
            reconnect?: boolean | undefined;
            maxReconnectAttempts?: number | undefined;
            fallbackServers?: string[] | undefined;
        } | undefined;
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    version: number;
    servers: {
        name: string;
        id: string;
        enabled: boolean;
        transport: {
            type: "stdio";
            command: string;
            args: string[];
            env: Record<string, string>;
            cwd?: string | undefined;
        } | {
            type: "sse";
            url: string;
            headers: Record<string, string>;
        } | {
            type: "websocket";
            url: string;
            headers: Record<string, string>;
        };
        source: "manual" | "curated" | "marketplace";
        settings: {
            autoConnect: boolean;
            timeout: number;
            reconnect: boolean;
            maxReconnectAttempts: number;
            fallbackServers: string[];
        };
        status?: {
            connected: boolean;
            reconnectAttempts: number;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
        } | undefined;
        capabilities?: {
            tools: {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }[];
            resources: {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }[];
            prompts: {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required: boolean;
                }[] | undefined;
            }[];
        } | undefined;
    }[];
}, {
    servers: {
        name: string;
        id: string;
        transport: {
            type: "stdio";
            command: string;
            args?: string[] | undefined;
            env?: Record<string, string> | undefined;
            cwd?: string | undefined;
        } | {
            type: "sse";
            url: string;
            headers?: Record<string, string> | undefined;
        } | {
            type: "websocket";
            url: string;
            headers?: Record<string, string> | undefined;
        };
        status?: {
            connected: boolean;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
            reconnectAttempts?: number | undefined;
        } | undefined;
        enabled?: boolean | undefined;
        capabilities?: {
            tools?: {
                name: string;
                description: string;
                inputSchema: Record<string, unknown>;
            }[] | undefined;
            resources?: {
                name: string;
                uri: string;
                mimeType?: string | undefined;
            }[] | undefined;
            prompts?: {
                name: string;
                description: string;
                arguments?: {
                    name: string;
                    description: string;
                    required?: boolean | undefined;
                }[] | undefined;
            }[] | undefined;
        } | undefined;
        source?: "manual" | "curated" | "marketplace" | undefined;
        settings?: {
            autoConnect?: boolean | undefined;
            timeout?: number | undefined;
            reconnect?: boolean | undefined;
            maxReconnectAttempts?: number | undefined;
            fallbackServers?: string[] | undefined;
        } | undefined;
    }[];
    version?: number | undefined;
}>;
export declare const ValidationResultSchema: z.ZodObject<{
    valid: z.ZodBoolean;
    errors: z.ZodArray<z.ZodString, "many">;
}, "strip", z.ZodTypeAny, {
    valid: boolean;
    errors: string[];
}, {
    valid: boolean;
    errors: string[];
}>;
export declare const ConnectionTestResultSchema: z.ZodObject<{
    success: z.ZodBoolean;
    latency: z.ZodOptional<z.ZodNumber>;
    error: z.ZodOptional<z.ZodString>;
    capabilities: z.ZodOptional<z.ZodObject<{
        tools: z.ZodDefault<z.ZodArray<z.ZodObject<{
            name: z.ZodString;
            description: z.ZodString;
            inputSchema: z.ZodRecord<z.ZodString, z.ZodUnknown>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }, {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }>, "many">>;
        resources: z.ZodDefault<z.ZodArray<z.ZodObject<{
            uri: z.ZodString;
            name: z.ZodString;
            mimeType: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }, {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }>, "many">>;
        prompts: z.ZodDefault<z.ZodArray<z.ZodObject<{
            name: z.ZodString;
            description: z.ZodString;
            arguments: z.ZodOptional<z.ZodArray<z.ZodObject<{
                name: z.ZodString;
                description: z.ZodString;
                required: z.ZodDefault<z.ZodBoolean>;
            }, "strip", z.ZodTypeAny, {
                name: string;
                description: string;
                required: boolean;
            }, {
                name: string;
                description: string;
                required?: boolean | undefined;
            }>, "many">>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required: boolean;
            }[] | undefined;
        }, {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required?: boolean | undefined;
            }[] | undefined;
        }>, "many">>;
    }, "strip", z.ZodTypeAny, {
        tools: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[];
        resources: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[];
        prompts: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required: boolean;
            }[] | undefined;
        }[];
    }, {
        tools?: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[] | undefined;
        resources?: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[] | undefined;
        prompts?: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required?: boolean | undefined;
            }[] | undefined;
        }[] | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    success: boolean;
    capabilities?: {
        tools: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[];
        resources: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[];
        prompts: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required: boolean;
            }[] | undefined;
        }[];
    } | undefined;
    latency?: number | undefined;
    error?: string | undefined;
}, {
    success: boolean;
    capabilities?: {
        tools?: {
            name: string;
            description: string;
            inputSchema: Record<string, unknown>;
        }[] | undefined;
        resources?: {
            name: string;
            uri: string;
            mimeType?: string | undefined;
        }[] | undefined;
        prompts?: {
            name: string;
            description: string;
            arguments?: {
                name: string;
                description: string;
                required?: boolean | undefined;
            }[] | undefined;
        }[] | undefined;
    } | undefined;
    latency?: number | undefined;
    error?: string | undefined;
}>;
export type TransportType = z.infer<typeof TransportType>;
export type ServerSource = z.infer<typeof ServerSource>;
export type ToolDefinition = z.infer<typeof ToolDefinitionSchema>;
export type ResourceDefinition = z.infer<typeof ResourceDefinitionSchema>;
export type ArgumentDefinition = z.infer<typeof ArgumentDefinitionSchema>;
export type PromptDefinition = z.infer<typeof PromptDefinitionSchema>;
export type Capabilities = z.infer<typeof CapabilitiesSchema>;
export type StdioTransport = z.infer<typeof StdioTransportSchema>;
export type SSETransport = z.infer<typeof SSETransportSchema>;
export type WebSocketTransport = z.infer<typeof WebSocketTransportSchema>;
export type TransportConfig = z.infer<typeof TransportConfigSchema>;
export type ServerSettings = z.infer<typeof ServerSettingsSchema>;
export type ConnectionStatus = z.infer<typeof ConnectionStatusSchema>;
export type MCPServerConfig = z.infer<typeof MCPServerConfigSchema>;
export type MCPServersConfig = z.infer<typeof MCPServersConfigSchema>;
export type ValidationResult = z.infer<typeof ValidationResultSchema>;
export type ConnectionTestResult = z.infer<typeof ConnectionTestResultSchema>;
export declare class MCPError extends Error {
    constructor(message: string);
}
export declare class MCPServerNotFoundError extends MCPError {
    constructor(id: string);
}
export declare class MCPValidationError extends MCPError {
    errors: string[];
    constructor(errors: string[]);
}
export declare class MCPServerAlreadyExistsError extends MCPError {
    constructor(id: string);
}
export declare const CURRENT_VERSION = 1;
//# sourceMappingURL=types.d.ts.map