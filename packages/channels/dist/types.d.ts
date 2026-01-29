import { z } from 'zod';
export declare const ChannelType: z.ZodEnum<["telegram", "discord", "slack", "email", "whatsapp", "webhook"]>;
export declare const RateLimitConfigSchema: z.ZodObject<{
    requestsPerMinute: z.ZodOptional<z.ZodNumber>;
    tokensPerMinute: z.ZodOptional<z.ZodNumber>;
    requestsPerDay: z.ZodOptional<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    requestsPerMinute?: number | undefined;
    tokensPerMinute?: number | undefined;
    requestsPerDay?: number | undefined;
}, {
    requestsPerMinute?: number | undefined;
    tokensPerMinute?: number | undefined;
    requestsPerDay?: number | undefined;
}>;
export declare const TelegramConfigSchema: z.ZodObject<{
    botToken: z.ZodString;
    chatId: z.ZodOptional<z.ZodString>;
    parseMode: z.ZodOptional<z.ZodEnum<["HTML", "Markdown", "MarkdownV2"]>>;
    disableNotification: z.ZodOptional<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    botToken: string;
    chatId?: string | undefined;
    parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
    disableNotification?: boolean | undefined;
}, {
    botToken: string;
    chatId?: string | undefined;
    parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
    disableNotification?: boolean | undefined;
}>;
export declare const DiscordConfigSchema: z.ZodObject<{
    botToken: z.ZodString;
    applicationId: z.ZodString;
    guildId: z.ZodOptional<z.ZodString>;
    channelId: z.ZodOptional<z.ZodString>;
    intents: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
}, "strip", z.ZodTypeAny, {
    botToken: string;
    applicationId: string;
    intents: string[];
    guildId?: string | undefined;
    channelId?: string | undefined;
}, {
    botToken: string;
    applicationId: string;
    guildId?: string | undefined;
    channelId?: string | undefined;
    intents?: string[] | undefined;
}>;
export declare const SlackConfigSchema: z.ZodObject<{
    botToken: z.ZodString;
    appToken: z.ZodOptional<z.ZodString>;
    signingSecret: z.ZodOptional<z.ZodString>;
    channelId: z.ZodOptional<z.ZodString>;
    socketMode: z.ZodDefault<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    botToken: string;
    socketMode: boolean;
    channelId?: string | undefined;
    appToken?: string | undefined;
    signingSecret?: string | undefined;
}, {
    botToken: string;
    channelId?: string | undefined;
    appToken?: string | undefined;
    signingSecret?: string | undefined;
    socketMode?: boolean | undefined;
}>;
export declare const EmailServerConfigSchema: z.ZodObject<{
    host: z.ZodString;
    port: z.ZodNumber;
    secure: z.ZodBoolean;
    username: z.ZodString;
    password: z.ZodString;
}, "strip", z.ZodTypeAny, {
    host: string;
    port: number;
    secure: boolean;
    username: string;
    password: string;
}, {
    host: string;
    port: number;
    secure: boolean;
    username: string;
    password: string;
}>;
export declare const EmailConfigSchema: z.ZodObject<{
    imap: z.ZodOptional<z.ZodObject<{
        host: z.ZodString;
        port: z.ZodNumber;
        secure: z.ZodBoolean;
        username: z.ZodString;
        password: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    }, {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    }>>;
    smtp: z.ZodOptional<z.ZodObject<{
        host: z.ZodString;
        port: z.ZodNumber;
        secure: z.ZodBoolean;
        username: z.ZodString;
        password: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    }, {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    }>>;
    checkInterval: z.ZodDefault<z.ZodNumber>;
    markAsRead: z.ZodDefault<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    checkInterval: number;
    markAsRead: boolean;
    imap?: {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    } | undefined;
    smtp?: {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    } | undefined;
}, {
    imap?: {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    } | undefined;
    smtp?: {
        host: string;
        port: number;
        secure: boolean;
        username: string;
        password: string;
    } | undefined;
    checkInterval?: number | undefined;
    markAsRead?: boolean | undefined;
}>;
export declare const WhatsAppConfigSchema: z.ZodObject<{
    sessionName: z.ZodString;
    phoneNumber: z.ZodOptional<z.ZodString>;
    qrTimeout: z.ZodDefault<z.ZodNumber>;
    pairingCode: z.ZodDefault<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    sessionName: string;
    qrTimeout: number;
    pairingCode: boolean;
    phoneNumber?: string | undefined;
}, {
    sessionName: string;
    phoneNumber?: string | undefined;
    qrTimeout?: number | undefined;
    pairingCode?: boolean | undefined;
}>;
export declare const WebhookAuthSchema: z.ZodObject<{
    type: z.ZodEnum<["bearer", "basic", "api-key"]>;
    token: z.ZodOptional<z.ZodString>;
    username: z.ZodOptional<z.ZodString>;
    password: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    type: "bearer" | "basic" | "api-key";
    username?: string | undefined;
    password?: string | undefined;
    token?: string | undefined;
}, {
    type: "bearer" | "basic" | "api-key";
    username?: string | undefined;
    password?: string | undefined;
    token?: string | undefined;
}>;
export declare const WebhookRetryPolicySchema: z.ZodObject<{
    maxRetries: z.ZodDefault<z.ZodNumber>;
    backoffMs: z.ZodDefault<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    maxRetries: number;
    backoffMs: number;
}, {
    maxRetries?: number | undefined;
    backoffMs?: number | undefined;
}>;
export declare const WebhookConfigSchema: z.ZodObject<{
    url: z.ZodString;
    method: z.ZodDefault<z.ZodEnum<["GET", "POST", "PUT", "DELETE"]>>;
    headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
    auth: z.ZodOptional<z.ZodObject<{
        type: z.ZodEnum<["bearer", "basic", "api-key"]>;
        token: z.ZodOptional<z.ZodString>;
        username: z.ZodOptional<z.ZodString>;
        password: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        type: "bearer" | "basic" | "api-key";
        username?: string | undefined;
        password?: string | undefined;
        token?: string | undefined;
    }, {
        type: "bearer" | "basic" | "api-key";
        username?: string | undefined;
        password?: string | undefined;
        token?: string | undefined;
    }>>;
    retryPolicy: z.ZodDefault<z.ZodObject<{
        maxRetries: z.ZodDefault<z.ZodNumber>;
        backoffMs: z.ZodDefault<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        maxRetries: number;
        backoffMs: number;
    }, {
        maxRetries?: number | undefined;
        backoffMs?: number | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    url: string;
    method: "GET" | "POST" | "PUT" | "DELETE";
    headers: Record<string, string>;
    retryPolicy: {
        maxRetries: number;
        backoffMs: number;
    };
    auth?: {
        type: "bearer" | "basic" | "api-key";
        username?: string | undefined;
        password?: string | undefined;
        token?: string | undefined;
    } | undefined;
}, {
    url: string;
    method?: "GET" | "POST" | "PUT" | "DELETE" | undefined;
    headers?: Record<string, string> | undefined;
    auth?: {
        type: "bearer" | "basic" | "api-key";
        username?: string | undefined;
        password?: string | undefined;
        token?: string | undefined;
    } | undefined;
    retryPolicy?: {
        maxRetries?: number | undefined;
        backoffMs?: number | undefined;
    } | undefined;
}>;
export declare const ChannelSettingsSchema: z.ZodObject<{
    allowCommands: z.ZodDefault<z.ZodBoolean>;
    autoReply: z.ZodDefault<z.ZodBoolean>;
    filterPatterns: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    allowedUsers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    blockedUsers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    rateLimit: z.ZodOptional<z.ZodObject<{
        requestsPerMinute: z.ZodOptional<z.ZodNumber>;
        tokensPerMinute: z.ZodOptional<z.ZodNumber>;
        requestsPerDay: z.ZodOptional<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    }, {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    allowCommands: boolean;
    autoReply: boolean;
    filterPatterns: string[];
    allowedUsers: string[];
    blockedUsers: string[];
    rateLimit?: {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    } | undefined;
}, {
    allowCommands?: boolean | undefined;
    autoReply?: boolean | undefined;
    filterPatterns?: string[] | undefined;
    allowedUsers?: string[] | undefined;
    blockedUsers?: string[] | undefined;
    rateLimit?: {
        requestsPerMinute?: number | undefined;
        tokensPerMinute?: number | undefined;
        requestsPerDay?: number | undefined;
    } | undefined;
}>;
export declare const WebhookSettingsSchema: z.ZodObject<{
    url: z.ZodString;
    secret: z.ZodOptional<z.ZodString>;
    enabled: z.ZodDefault<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    url: string;
    enabled: boolean;
    secret?: string | undefined;
}, {
    url: string;
    secret?: string | undefined;
    enabled?: boolean | undefined;
}>;
export declare const ConnectionStatusSchema: z.ZodObject<{
    connected: z.ZodBoolean;
    lastConnected: z.ZodOptional<z.ZodString>;
    lastError: z.ZodOptional<z.ZodString>;
    errorCount: z.ZodDefault<z.ZodNumber>;
    messageCount: z.ZodDefault<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    connected: boolean;
    errorCount: number;
    messageCount: number;
    lastConnected?: string | undefined;
    lastError?: string | undefined;
}, {
    connected: boolean;
    lastConnected?: string | undefined;
    lastError?: string | undefined;
    errorCount?: number | undefined;
    messageCount?: number | undefined;
}>;
export declare const ChannelConfigSchema: z.ZodObject<{
    id: z.ZodString;
    name: z.ZodString;
    type: z.ZodEnum<["telegram", "discord", "slack", "email", "whatsapp", "webhook"]>;
    enabled: z.ZodDefault<z.ZodBoolean>;
    config: z.ZodUnion<[z.ZodObject<{
        botToken: z.ZodString;
        chatId: z.ZodOptional<z.ZodString>;
        parseMode: z.ZodOptional<z.ZodEnum<["HTML", "Markdown", "MarkdownV2"]>>;
        disableNotification: z.ZodOptional<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        botToken: string;
        chatId?: string | undefined;
        parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
        disableNotification?: boolean | undefined;
    }, {
        botToken: string;
        chatId?: string | undefined;
        parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
        disableNotification?: boolean | undefined;
    }>, z.ZodObject<{
        botToken: z.ZodString;
        applicationId: z.ZodString;
        guildId: z.ZodOptional<z.ZodString>;
        channelId: z.ZodOptional<z.ZodString>;
        intents: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    }, "strip", z.ZodTypeAny, {
        botToken: string;
        applicationId: string;
        intents: string[];
        guildId?: string | undefined;
        channelId?: string | undefined;
    }, {
        botToken: string;
        applicationId: string;
        guildId?: string | undefined;
        channelId?: string | undefined;
        intents?: string[] | undefined;
    }>, z.ZodObject<{
        botToken: z.ZodString;
        appToken: z.ZodOptional<z.ZodString>;
        signingSecret: z.ZodOptional<z.ZodString>;
        channelId: z.ZodOptional<z.ZodString>;
        socketMode: z.ZodDefault<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        botToken: string;
        socketMode: boolean;
        channelId?: string | undefined;
        appToken?: string | undefined;
        signingSecret?: string | undefined;
    }, {
        botToken: string;
        channelId?: string | undefined;
        appToken?: string | undefined;
        signingSecret?: string | undefined;
        socketMode?: boolean | undefined;
    }>, z.ZodObject<{
        imap: z.ZodOptional<z.ZodObject<{
            host: z.ZodString;
            port: z.ZodNumber;
            secure: z.ZodBoolean;
            username: z.ZodString;
            password: z.ZodString;
        }, "strip", z.ZodTypeAny, {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        }, {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        }>>;
        smtp: z.ZodOptional<z.ZodObject<{
            host: z.ZodString;
            port: z.ZodNumber;
            secure: z.ZodBoolean;
            username: z.ZodString;
            password: z.ZodString;
        }, "strip", z.ZodTypeAny, {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        }, {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        }>>;
        checkInterval: z.ZodDefault<z.ZodNumber>;
        markAsRead: z.ZodDefault<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        checkInterval: number;
        markAsRead: boolean;
        imap?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
        smtp?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
    }, {
        imap?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
        smtp?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
        checkInterval?: number | undefined;
        markAsRead?: boolean | undefined;
    }>, z.ZodObject<{
        sessionName: z.ZodString;
        phoneNumber: z.ZodOptional<z.ZodString>;
        qrTimeout: z.ZodDefault<z.ZodNumber>;
        pairingCode: z.ZodDefault<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        sessionName: string;
        qrTimeout: number;
        pairingCode: boolean;
        phoneNumber?: string | undefined;
    }, {
        sessionName: string;
        phoneNumber?: string | undefined;
        qrTimeout?: number | undefined;
        pairingCode?: boolean | undefined;
    }>, z.ZodObject<{
        url: z.ZodString;
        method: z.ZodDefault<z.ZodEnum<["GET", "POST", "PUT", "DELETE"]>>;
        headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
        auth: z.ZodOptional<z.ZodObject<{
            type: z.ZodEnum<["bearer", "basic", "api-key"]>;
            token: z.ZodOptional<z.ZodString>;
            username: z.ZodOptional<z.ZodString>;
            password: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            type: "bearer" | "basic" | "api-key";
            username?: string | undefined;
            password?: string | undefined;
            token?: string | undefined;
        }, {
            type: "bearer" | "basic" | "api-key";
            username?: string | undefined;
            password?: string | undefined;
            token?: string | undefined;
        }>>;
        retryPolicy: z.ZodDefault<z.ZodObject<{
            maxRetries: z.ZodDefault<z.ZodNumber>;
            backoffMs: z.ZodDefault<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            maxRetries: number;
            backoffMs: number;
        }, {
            maxRetries?: number | undefined;
            backoffMs?: number | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        url: string;
        method: "GET" | "POST" | "PUT" | "DELETE";
        headers: Record<string, string>;
        retryPolicy: {
            maxRetries: number;
            backoffMs: number;
        };
        auth?: {
            type: "bearer" | "basic" | "api-key";
            username?: string | undefined;
            password?: string | undefined;
            token?: string | undefined;
        } | undefined;
    }, {
        url: string;
        method?: "GET" | "POST" | "PUT" | "DELETE" | undefined;
        headers?: Record<string, string> | undefined;
        auth?: {
            type: "bearer" | "basic" | "api-key";
            username?: string | undefined;
            password?: string | undefined;
            token?: string | undefined;
        } | undefined;
        retryPolicy?: {
            maxRetries?: number | undefined;
            backoffMs?: number | undefined;
        } | undefined;
    }>]>;
    settings: z.ZodDefault<z.ZodObject<{
        allowCommands: z.ZodDefault<z.ZodBoolean>;
        autoReply: z.ZodDefault<z.ZodBoolean>;
        filterPatterns: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        allowedUsers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        blockedUsers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        rateLimit: z.ZodOptional<z.ZodObject<{
            requestsPerMinute: z.ZodOptional<z.ZodNumber>;
            tokensPerMinute: z.ZodOptional<z.ZodNumber>;
            requestsPerDay: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        }, {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        allowCommands: boolean;
        autoReply: boolean;
        filterPatterns: string[];
        allowedUsers: string[];
        blockedUsers: string[];
        rateLimit?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
    }, {
        allowCommands?: boolean | undefined;
        autoReply?: boolean | undefined;
        filterPatterns?: string[] | undefined;
        allowedUsers?: string[] | undefined;
        blockedUsers?: string[] | undefined;
        rateLimit?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
    }>>;
    webhook: z.ZodOptional<z.ZodObject<{
        url: z.ZodString;
        secret: z.ZodOptional<z.ZodString>;
        enabled: z.ZodDefault<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        url: string;
        enabled: boolean;
        secret?: string | undefined;
    }, {
        url: string;
        secret?: string | undefined;
        enabled?: boolean | undefined;
    }>>;
    status: z.ZodOptional<z.ZodObject<{
        connected: z.ZodBoolean;
        lastConnected: z.ZodOptional<z.ZodString>;
        lastError: z.ZodOptional<z.ZodString>;
        errorCount: z.ZodDefault<z.ZodNumber>;
        messageCount: z.ZodDefault<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        connected: boolean;
        errorCount: number;
        messageCount: number;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
    }, {
        connected: boolean;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
        errorCount?: number | undefined;
        messageCount?: number | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    type: "telegram" | "discord" | "slack" | "email" | "whatsapp" | "webhook";
    enabled: boolean;
    id: string;
    name: string;
    config: {
        botToken: string;
        chatId?: string | undefined;
        parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
        disableNotification?: boolean | undefined;
    } | {
        botToken: string;
        applicationId: string;
        intents: string[];
        guildId?: string | undefined;
        channelId?: string | undefined;
    } | {
        botToken: string;
        socketMode: boolean;
        channelId?: string | undefined;
        appToken?: string | undefined;
        signingSecret?: string | undefined;
    } | {
        checkInterval: number;
        markAsRead: boolean;
        imap?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
        smtp?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
    } | {
        sessionName: string;
        qrTimeout: number;
        pairingCode: boolean;
        phoneNumber?: string | undefined;
    } | {
        url: string;
        method: "GET" | "POST" | "PUT" | "DELETE";
        headers: Record<string, string>;
        retryPolicy: {
            maxRetries: number;
            backoffMs: number;
        };
        auth?: {
            type: "bearer" | "basic" | "api-key";
            username?: string | undefined;
            password?: string | undefined;
            token?: string | undefined;
        } | undefined;
    };
    settings: {
        allowCommands: boolean;
        autoReply: boolean;
        filterPatterns: string[];
        allowedUsers: string[];
        blockedUsers: string[];
        rateLimit?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
    };
    webhook?: {
        url: string;
        enabled: boolean;
        secret?: string | undefined;
    } | undefined;
    status?: {
        connected: boolean;
        errorCount: number;
        messageCount: number;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
    } | undefined;
}, {
    type: "telegram" | "discord" | "slack" | "email" | "whatsapp" | "webhook";
    id: string;
    name: string;
    config: {
        botToken: string;
        chatId?: string | undefined;
        parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
        disableNotification?: boolean | undefined;
    } | {
        botToken: string;
        applicationId: string;
        guildId?: string | undefined;
        channelId?: string | undefined;
        intents?: string[] | undefined;
    } | {
        botToken: string;
        channelId?: string | undefined;
        appToken?: string | undefined;
        signingSecret?: string | undefined;
        socketMode?: boolean | undefined;
    } | {
        imap?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
        smtp?: {
            host: string;
            port: number;
            secure: boolean;
            username: string;
            password: string;
        } | undefined;
        checkInterval?: number | undefined;
        markAsRead?: boolean | undefined;
    } | {
        sessionName: string;
        phoneNumber?: string | undefined;
        qrTimeout?: number | undefined;
        pairingCode?: boolean | undefined;
    } | {
        url: string;
        method?: "GET" | "POST" | "PUT" | "DELETE" | undefined;
        headers?: Record<string, string> | undefined;
        auth?: {
            type: "bearer" | "basic" | "api-key";
            username?: string | undefined;
            password?: string | undefined;
            token?: string | undefined;
        } | undefined;
        retryPolicy?: {
            maxRetries?: number | undefined;
            backoffMs?: number | undefined;
        } | undefined;
    };
    webhook?: {
        url: string;
        secret?: string | undefined;
        enabled?: boolean | undefined;
    } | undefined;
    status?: {
        connected: boolean;
        lastConnected?: string | undefined;
        lastError?: string | undefined;
        errorCount?: number | undefined;
        messageCount?: number | undefined;
    } | undefined;
    enabled?: boolean | undefined;
    settings?: {
        allowCommands?: boolean | undefined;
        autoReply?: boolean | undefined;
        filterPatterns?: string[] | undefined;
        allowedUsers?: string[] | undefined;
        blockedUsers?: string[] | undefined;
        rateLimit?: {
            requestsPerMinute?: number | undefined;
            tokensPerMinute?: number | undefined;
            requestsPerDay?: number | undefined;
        } | undefined;
    } | undefined;
}>;
export declare const ChannelsConfigSchema: z.ZodObject<{
    version: z.ZodDefault<z.ZodNumber>;
    channels: z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        name: z.ZodString;
        type: z.ZodEnum<["telegram", "discord", "slack", "email", "whatsapp", "webhook"]>;
        enabled: z.ZodDefault<z.ZodBoolean>;
        config: z.ZodUnion<[z.ZodObject<{
            botToken: z.ZodString;
            chatId: z.ZodOptional<z.ZodString>;
            parseMode: z.ZodOptional<z.ZodEnum<["HTML", "Markdown", "MarkdownV2"]>>;
            disableNotification: z.ZodOptional<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            botToken: string;
            chatId?: string | undefined;
            parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
            disableNotification?: boolean | undefined;
        }, {
            botToken: string;
            chatId?: string | undefined;
            parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
            disableNotification?: boolean | undefined;
        }>, z.ZodObject<{
            botToken: z.ZodString;
            applicationId: z.ZodString;
            guildId: z.ZodOptional<z.ZodString>;
            channelId: z.ZodOptional<z.ZodString>;
            intents: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        }, "strip", z.ZodTypeAny, {
            botToken: string;
            applicationId: string;
            intents: string[];
            guildId?: string | undefined;
            channelId?: string | undefined;
        }, {
            botToken: string;
            applicationId: string;
            guildId?: string | undefined;
            channelId?: string | undefined;
            intents?: string[] | undefined;
        }>, z.ZodObject<{
            botToken: z.ZodString;
            appToken: z.ZodOptional<z.ZodString>;
            signingSecret: z.ZodOptional<z.ZodString>;
            channelId: z.ZodOptional<z.ZodString>;
            socketMode: z.ZodDefault<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            botToken: string;
            socketMode: boolean;
            channelId?: string | undefined;
            appToken?: string | undefined;
            signingSecret?: string | undefined;
        }, {
            botToken: string;
            channelId?: string | undefined;
            appToken?: string | undefined;
            signingSecret?: string | undefined;
            socketMode?: boolean | undefined;
        }>, z.ZodObject<{
            imap: z.ZodOptional<z.ZodObject<{
                host: z.ZodString;
                port: z.ZodNumber;
                secure: z.ZodBoolean;
                username: z.ZodString;
                password: z.ZodString;
            }, "strip", z.ZodTypeAny, {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            }, {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            }>>;
            smtp: z.ZodOptional<z.ZodObject<{
                host: z.ZodString;
                port: z.ZodNumber;
                secure: z.ZodBoolean;
                username: z.ZodString;
                password: z.ZodString;
            }, "strip", z.ZodTypeAny, {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            }, {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            }>>;
            checkInterval: z.ZodDefault<z.ZodNumber>;
            markAsRead: z.ZodDefault<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            checkInterval: number;
            markAsRead: boolean;
            imap?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            smtp?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
        }, {
            imap?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            smtp?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            checkInterval?: number | undefined;
            markAsRead?: boolean | undefined;
        }>, z.ZodObject<{
            sessionName: z.ZodString;
            phoneNumber: z.ZodOptional<z.ZodString>;
            qrTimeout: z.ZodDefault<z.ZodNumber>;
            pairingCode: z.ZodDefault<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            sessionName: string;
            qrTimeout: number;
            pairingCode: boolean;
            phoneNumber?: string | undefined;
        }, {
            sessionName: string;
            phoneNumber?: string | undefined;
            qrTimeout?: number | undefined;
            pairingCode?: boolean | undefined;
        }>, z.ZodObject<{
            url: z.ZodString;
            method: z.ZodDefault<z.ZodEnum<["GET", "POST", "PUT", "DELETE"]>>;
            headers: z.ZodDefault<z.ZodRecord<z.ZodString, z.ZodString>>;
            auth: z.ZodOptional<z.ZodObject<{
                type: z.ZodEnum<["bearer", "basic", "api-key"]>;
                token: z.ZodOptional<z.ZodString>;
                username: z.ZodOptional<z.ZodString>;
                password: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            }, {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            }>>;
            retryPolicy: z.ZodDefault<z.ZodObject<{
                maxRetries: z.ZodDefault<z.ZodNumber>;
                backoffMs: z.ZodDefault<z.ZodNumber>;
            }, "strip", z.ZodTypeAny, {
                maxRetries: number;
                backoffMs: number;
            }, {
                maxRetries?: number | undefined;
                backoffMs?: number | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            url: string;
            method: "GET" | "POST" | "PUT" | "DELETE";
            headers: Record<string, string>;
            retryPolicy: {
                maxRetries: number;
                backoffMs: number;
            };
            auth?: {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            } | undefined;
        }, {
            url: string;
            method?: "GET" | "POST" | "PUT" | "DELETE" | undefined;
            headers?: Record<string, string> | undefined;
            auth?: {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            } | undefined;
            retryPolicy?: {
                maxRetries?: number | undefined;
                backoffMs?: number | undefined;
            } | undefined;
        }>]>;
        settings: z.ZodDefault<z.ZodObject<{
            allowCommands: z.ZodDefault<z.ZodBoolean>;
            autoReply: z.ZodDefault<z.ZodBoolean>;
            filterPatterns: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
            allowedUsers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
            blockedUsers: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
            rateLimit: z.ZodOptional<z.ZodObject<{
                requestsPerMinute: z.ZodOptional<z.ZodNumber>;
                tokensPerMinute: z.ZodOptional<z.ZodNumber>;
                requestsPerDay: z.ZodOptional<z.ZodNumber>;
            }, "strip", z.ZodTypeAny, {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            }, {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            allowCommands: boolean;
            autoReply: boolean;
            filterPatterns: string[];
            allowedUsers: string[];
            blockedUsers: string[];
            rateLimit?: {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            } | undefined;
        }, {
            allowCommands?: boolean | undefined;
            autoReply?: boolean | undefined;
            filterPatterns?: string[] | undefined;
            allowedUsers?: string[] | undefined;
            blockedUsers?: string[] | undefined;
            rateLimit?: {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            } | undefined;
        }>>;
        webhook: z.ZodOptional<z.ZodObject<{
            url: z.ZodString;
            secret: z.ZodOptional<z.ZodString>;
            enabled: z.ZodDefault<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            url: string;
            enabled: boolean;
            secret?: string | undefined;
        }, {
            url: string;
            secret?: string | undefined;
            enabled?: boolean | undefined;
        }>>;
        status: z.ZodOptional<z.ZodObject<{
            connected: z.ZodBoolean;
            lastConnected: z.ZodOptional<z.ZodString>;
            lastError: z.ZodOptional<z.ZodString>;
            errorCount: z.ZodDefault<z.ZodNumber>;
            messageCount: z.ZodDefault<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            connected: boolean;
            errorCount: number;
            messageCount: number;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
        }, {
            connected: boolean;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
            errorCount?: number | undefined;
            messageCount?: number | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        type: "telegram" | "discord" | "slack" | "email" | "whatsapp" | "webhook";
        enabled: boolean;
        id: string;
        name: string;
        config: {
            botToken: string;
            chatId?: string | undefined;
            parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
            disableNotification?: boolean | undefined;
        } | {
            botToken: string;
            applicationId: string;
            intents: string[];
            guildId?: string | undefined;
            channelId?: string | undefined;
        } | {
            botToken: string;
            socketMode: boolean;
            channelId?: string | undefined;
            appToken?: string | undefined;
            signingSecret?: string | undefined;
        } | {
            checkInterval: number;
            markAsRead: boolean;
            imap?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            smtp?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
        } | {
            sessionName: string;
            qrTimeout: number;
            pairingCode: boolean;
            phoneNumber?: string | undefined;
        } | {
            url: string;
            method: "GET" | "POST" | "PUT" | "DELETE";
            headers: Record<string, string>;
            retryPolicy: {
                maxRetries: number;
                backoffMs: number;
            };
            auth?: {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            } | undefined;
        };
        settings: {
            allowCommands: boolean;
            autoReply: boolean;
            filterPatterns: string[];
            allowedUsers: string[];
            blockedUsers: string[];
            rateLimit?: {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            } | undefined;
        };
        webhook?: {
            url: string;
            enabled: boolean;
            secret?: string | undefined;
        } | undefined;
        status?: {
            connected: boolean;
            errorCount: number;
            messageCount: number;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
        } | undefined;
    }, {
        type: "telegram" | "discord" | "slack" | "email" | "whatsapp" | "webhook";
        id: string;
        name: string;
        config: {
            botToken: string;
            chatId?: string | undefined;
            parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
            disableNotification?: boolean | undefined;
        } | {
            botToken: string;
            applicationId: string;
            guildId?: string | undefined;
            channelId?: string | undefined;
            intents?: string[] | undefined;
        } | {
            botToken: string;
            channelId?: string | undefined;
            appToken?: string | undefined;
            signingSecret?: string | undefined;
            socketMode?: boolean | undefined;
        } | {
            imap?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            smtp?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            checkInterval?: number | undefined;
            markAsRead?: boolean | undefined;
        } | {
            sessionName: string;
            phoneNumber?: string | undefined;
            qrTimeout?: number | undefined;
            pairingCode?: boolean | undefined;
        } | {
            url: string;
            method?: "GET" | "POST" | "PUT" | "DELETE" | undefined;
            headers?: Record<string, string> | undefined;
            auth?: {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            } | undefined;
            retryPolicy?: {
                maxRetries?: number | undefined;
                backoffMs?: number | undefined;
            } | undefined;
        };
        webhook?: {
            url: string;
            secret?: string | undefined;
            enabled?: boolean | undefined;
        } | undefined;
        status?: {
            connected: boolean;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
            errorCount?: number | undefined;
            messageCount?: number | undefined;
        } | undefined;
        enabled?: boolean | undefined;
        settings?: {
            allowCommands?: boolean | undefined;
            autoReply?: boolean | undefined;
            filterPatterns?: string[] | undefined;
            allowedUsers?: string[] | undefined;
            blockedUsers?: string[] | undefined;
            rateLimit?: {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            } | undefined;
        } | undefined;
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    version: number;
    channels: {
        type: "telegram" | "discord" | "slack" | "email" | "whatsapp" | "webhook";
        enabled: boolean;
        id: string;
        name: string;
        config: {
            botToken: string;
            chatId?: string | undefined;
            parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
            disableNotification?: boolean | undefined;
        } | {
            botToken: string;
            applicationId: string;
            intents: string[];
            guildId?: string | undefined;
            channelId?: string | undefined;
        } | {
            botToken: string;
            socketMode: boolean;
            channelId?: string | undefined;
            appToken?: string | undefined;
            signingSecret?: string | undefined;
        } | {
            checkInterval: number;
            markAsRead: boolean;
            imap?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            smtp?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
        } | {
            sessionName: string;
            qrTimeout: number;
            pairingCode: boolean;
            phoneNumber?: string | undefined;
        } | {
            url: string;
            method: "GET" | "POST" | "PUT" | "DELETE";
            headers: Record<string, string>;
            retryPolicy: {
                maxRetries: number;
                backoffMs: number;
            };
            auth?: {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            } | undefined;
        };
        settings: {
            allowCommands: boolean;
            autoReply: boolean;
            filterPatterns: string[];
            allowedUsers: string[];
            blockedUsers: string[];
            rateLimit?: {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            } | undefined;
        };
        webhook?: {
            url: string;
            enabled: boolean;
            secret?: string | undefined;
        } | undefined;
        status?: {
            connected: boolean;
            errorCount: number;
            messageCount: number;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
        } | undefined;
    }[];
}, {
    channels: {
        type: "telegram" | "discord" | "slack" | "email" | "whatsapp" | "webhook";
        id: string;
        name: string;
        config: {
            botToken: string;
            chatId?: string | undefined;
            parseMode?: "HTML" | "Markdown" | "MarkdownV2" | undefined;
            disableNotification?: boolean | undefined;
        } | {
            botToken: string;
            applicationId: string;
            guildId?: string | undefined;
            channelId?: string | undefined;
            intents?: string[] | undefined;
        } | {
            botToken: string;
            channelId?: string | undefined;
            appToken?: string | undefined;
            signingSecret?: string | undefined;
            socketMode?: boolean | undefined;
        } | {
            imap?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            smtp?: {
                host: string;
                port: number;
                secure: boolean;
                username: string;
                password: string;
            } | undefined;
            checkInterval?: number | undefined;
            markAsRead?: boolean | undefined;
        } | {
            sessionName: string;
            phoneNumber?: string | undefined;
            qrTimeout?: number | undefined;
            pairingCode?: boolean | undefined;
        } | {
            url: string;
            method?: "GET" | "POST" | "PUT" | "DELETE" | undefined;
            headers?: Record<string, string> | undefined;
            auth?: {
                type: "bearer" | "basic" | "api-key";
                username?: string | undefined;
                password?: string | undefined;
                token?: string | undefined;
            } | undefined;
            retryPolicy?: {
                maxRetries?: number | undefined;
                backoffMs?: number | undefined;
            } | undefined;
        };
        webhook?: {
            url: string;
            secret?: string | undefined;
            enabled?: boolean | undefined;
        } | undefined;
        status?: {
            connected: boolean;
            lastConnected?: string | undefined;
            lastError?: string | undefined;
            errorCount?: number | undefined;
            messageCount?: number | undefined;
        } | undefined;
        enabled?: boolean | undefined;
        settings?: {
            allowCommands?: boolean | undefined;
            autoReply?: boolean | undefined;
            filterPatterns?: string[] | undefined;
            allowedUsers?: string[] | undefined;
            blockedUsers?: string[] | undefined;
            rateLimit?: {
                requestsPerMinute?: number | undefined;
                tokensPerMinute?: number | undefined;
                requestsPerDay?: number | undefined;
            } | undefined;
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
}, "strip", z.ZodTypeAny, {
    success: boolean;
    latency?: number | undefined;
    error?: string | undefined;
}, {
    success: boolean;
    latency?: number | undefined;
    error?: string | undefined;
}>;
export type ChannelType = z.infer<typeof ChannelType>;
export type RateLimitConfig = z.infer<typeof RateLimitConfigSchema>;
export type TelegramConfig = z.infer<typeof TelegramConfigSchema>;
export type DiscordConfig = z.infer<typeof DiscordConfigSchema>;
export type SlackConfig = z.infer<typeof SlackConfigSchema>;
export type EmailServerConfig = z.infer<typeof EmailServerConfigSchema>;
export type EmailConfig = z.infer<typeof EmailConfigSchema>;
export type WhatsAppConfig = z.infer<typeof WhatsAppConfigSchema>;
export type WebhookAuth = z.infer<typeof WebhookAuthSchema>;
export type WebhookRetryPolicy = z.infer<typeof WebhookRetryPolicySchema>;
export type WebhookConfig = z.infer<typeof WebhookConfigSchema>;
export type ChannelSettings = z.infer<typeof ChannelSettingsSchema>;
export type WebhookSettings = z.infer<typeof WebhookSettingsSchema>;
export type ConnectionStatus = z.infer<typeof ConnectionStatusSchema>;
export type ChannelConfig = z.infer<typeof ChannelConfigSchema>;
export type ChannelsConfig = z.infer<typeof ChannelsConfigSchema>;
export type ValidationResult = z.infer<typeof ValidationResultSchema>;
export type ConnectionTestResult = z.infer<typeof ConnectionTestResultSchema>;
export declare class ChannelError extends Error {
    constructor(message: string);
}
export declare class ChannelNotFoundError extends ChannelError {
    constructor(id: string);
}
export declare class ChannelValidationError extends ChannelError {
    errors: string[];
    constructor(errors: string[]);
}
export declare class ChannelAlreadyExistsError extends ChannelError {
    constructor(id: string);
}
export declare const CURRENT_VERSION = 1;
//# sourceMappingURL=types.d.ts.map