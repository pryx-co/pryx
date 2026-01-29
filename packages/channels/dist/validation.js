import { ChannelConfigSchema, ChannelValidationError, } from './types.js';
export function validateChannelConfig(config) {
    const result = ChannelConfigSchema.safeParse(config);
    if (result.success) {
        const errors = [];
        const validated = result.data;
        switch (validated.type) {
            case 'telegram': {
                if (!('botToken' in validated.config)) {
                    errors.push('Telegram config requires botToken');
                }
                break;
            }
            case 'discord': {
                if (!('botToken' in validated.config)) {
                    errors.push('Discord config requires botToken');
                }
                if (!('applicationId' in validated.config)) {
                    errors.push('Discord config requires applicationId');
                }
                break;
            }
            case 'slack': {
                if (!('botToken' in validated.config)) {
                    errors.push('Slack config requires botToken');
                }
                break;
            }
            case 'email': {
                const emailConfig = validated.config;
                if (!emailConfig.imap && !emailConfig.smtp) {
                    errors.push('Email config requires at least imap or smtp');
                }
                break;
            }
            case 'whatsapp': {
                if (!('sessionName' in validated.config)) {
                    errors.push('WhatsApp config requires sessionName');
                }
                break;
            }
            case 'webhook': {
                if (!('url' in validated.config)) {
                    errors.push('Webhook config requires url');
                }
                break;
            }
        }
        if (validated.webhook?.enabled && !validated.webhook.url) {
            errors.push('Webhook settings enabled but URL is missing');
        }
        return {
            valid: errors.length === 0,
            errors,
        };
    }
    return {
        valid: false,
        errors: result.error.errors.map((e) => `${e.path.join('.')}: ${e.message}`),
    };
}
export function assertValidChannelConfig(config) {
    const result = validateChannelConfig(config);
    if (!result.valid) {
        throw new ChannelValidationError(result.errors);
    }
    return config;
}
export function isValidChannelId(id) {
    return /^[a-z0-9_-]+$/.test(id) && id.length >= 1 && id.length <= 64;
}
export function isValidChannelType(type) {
    return ['telegram', 'discord', 'slack', 'email', 'whatsapp', 'webhook'].includes(type);
}
export function matchesFilterPatterns(message, patterns) {
    if (patterns.length === 0)
        return true;
    return patterns.some((pattern) => {
        try {
            const regex = new RegExp(pattern, 'i');
            return regex.test(message);
        }
        catch {
            return message.toLowerCase().includes(pattern.toLowerCase());
        }
    });
}
export function isUserAllowed(userId, allowedList, blockedList) {
    if (blockedList.includes(userId)) {
        return false;
    }
    if (allowedList.length > 0 && !allowedList.includes(userId)) {
        return false;
    }
    return true;
}
//# sourceMappingURL=validation.js.map