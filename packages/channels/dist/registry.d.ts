import { ChannelConfig, ChannelsConfig, ConnectionTestResult, ChannelType } from './types.js';
import { validateChannelConfig } from './validation.js';
export declare class ChannelRegistry {
    private _channels;
    private _version;
    addChannel(config: ChannelConfig): void;
    updateChannel(id: string, updates: Partial<ChannelConfig>): ChannelConfig;
    removeChannel(id: string): void;
    getChannel(id: string): ChannelConfig | undefined;
    getAllChannels(): ChannelConfig[];
    getEnabledChannels(): ChannelConfig[];
    getChannelsByType(type: ChannelType): ChannelConfig[];
    hasChannel(id: string): boolean;
    enableChannel(id: string): void;
    disableChannel(id: string): void;
    enableAll(): void;
    disableAll(): void;
    enableType(type: ChannelType): void;
    disableType(type: ChannelType): void;
    updateChannelStatus(id: string, status: Partial<ChannelConfig['status']>): void;
    validateChannel(id: string): ReturnType<typeof validateChannelConfig>;
    testConnection(id: string): Promise<ConnectionTestResult>;
    toJSON(): ChannelsConfig;
    fromJSON(data: ChannelsConfig): void;
    clear(): void;
    get size(): number;
}
export declare function createRegistry(): ChannelRegistry;
//# sourceMappingURL=registry.d.ts.map