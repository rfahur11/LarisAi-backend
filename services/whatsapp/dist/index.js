"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const cors_1 = __importDefault(require("cors"));
const baileys_1 = __importStar(require("@whiskeysockets/baileys"));
const qrcode = __importStar(require("qrcode-terminal"));
const pino_1 = __importDefault(require("pino"));
const app = (0, express_1.default)();
app.use((0, cors_1.default)());
app.use(express_1.default.json());
const PORT = process.env.PORT || 8002;
let sock = null;
async function connectToWhatsApp() {
    const { state, saveCreds } = await (0, baileys_1.useMultiFileAuthState)('auth_info_baileys');
    sock = (0, baileys_1.default)({
        auth: state,
        printQRInTerminal: true,
        logger: (0, pino_1.default)({ level: 'silent' })
    });
    sock.ev.on('connection.update', (update) => {
        const { connection, lastDisconnect, qr } = update;
        if (qr) {
            console.log('Please scan this QR code with your WhatsApp:');
            qrcode.generate(qr, { small: true });
        }
        if (connection === 'close') {
            const shouldReconnect = lastDisconnect?.error?.output?.statusCode !== baileys_1.DisconnectReason.loggedOut;
            console.log('Connection closed due to', lastDisconnect?.error, ', reconnecting', shouldReconnect);
            if (shouldReconnect) {
                connectToWhatsApp();
            }
        }
        else if (connection === 'open') {
            console.log('Opened connection to WhatsApp');
        }
    });
    sock.ev.on('creds.update', saveCreds);
}
// Endpoint to send message
app.post('/api/v1/whatsapp/send', async (req, res) => {
    try {
        const { to, message } = req.body;
        if (!to || !message) {
            return res.status(400).json({ success: false, error: 'Missing "to" or "message" in request body' });
        }
        if (!sock) {
            return res.status(500).json({ success: false, error: 'WhatsApp socket not initialized' });
        }
        // Format number correctly (add @s.whatsapp.net if missing)
        let jid = to;
        if (!jid.includes('@s.whatsapp.net')) {
            // Remove + or non-numeric characters from beginning
            let cleanNumber = jid.replace(/\D/g, '');
            // For Indonesian numbers starting with 0, replace with 62
            if (cleanNumber.startsWith('0')) {
                cleanNumber = '62' + cleanNumber.substring(1);
            }
            jid = `${cleanNumber}@s.whatsapp.net`;
        }
        const sentMsg = await sock.sendMessage(jid, { text: message });
        console.log(`Message sent to ${jid}`);
        return res.json({ success: true, messageId: sentMsg.key.id });
    }
    catch (error) {
        console.error('Error sending message:', error);
        return res.status(500).json({ success: false, error: error.message });
    }
});
// Start server
app.listen(PORT, () => {
    console.log(`WhatsApp Service listening on port ${PORT}`);
    // Initialize WhatsApp connection
    connectToWhatsApp();
});
