import express from 'express';
import cors from 'cors';
import makeWASocket, { DisconnectReason, useMultiFileAuthState } from '@whiskeysockets/baileys';
import * as qrcode from 'qrcode-terminal';
import { Boom } from '@hapi/boom';
import pino from 'pino';

const app = express();
app.use(cors());
app.use(express.json());

const PORT = process.env.PORT || 8002;
let sock: any = null;

async function connectToWhatsApp() {
    const { state, saveCreds } = await useMultiFileAuthState('auth_info_baileys');
    
    sock = makeWASocket({
        auth: state,
        printQRInTerminal: true,
        logger: pino({ level: 'silent' }) as any
    });

    sock.ev.on('connection.update', (update: any) => {
        const { connection, lastDisconnect, qr } = update;
        
        if (qr) {
            console.log('Please scan this QR code with your WhatsApp:');
            qrcode.generate(qr, { small: true });
        }

        if (connection === 'close') {
            const shouldReconnect = (lastDisconnect?.error as Boom)?.output?.statusCode !== DisconnectReason.loggedOut;
            console.log('Connection closed due to', lastDisconnect?.error, ', reconnecting', shouldReconnect);
            if (shouldReconnect) {
                connectToWhatsApp();
            }
        } else if (connection === 'open') {
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
    } catch (error: any) {
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
