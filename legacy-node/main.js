const readline = require('readline');
const { exec } = require('child_process');
const mysql = require('mysql2/promise');
const crypto = require('crypto');
const BigInteger = require('big-integer');

// 命令行交互
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

// ====================== RSA 工具类 ======================
class DataTranslate {
    static HexStringToByte(hex) {
        const len = hex.length / 2;
        const result = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
            const pos = i * 2;
            const high = parseInt(hex.substr(pos, 1), 16);
            const low = parseInt(hex.substr(pos + 1, 1), 16);
            result[i] = (high << 4) | low;
        }
        return Buffer.from(result);
    }
}

class RSAHelper {
    static ConvertFromPemPrivateKey(pem) {
        pem = pem.replace(/-----BEGIN RSA PRIVATE KEY-----/g, '').replace(/-----END RSA PRIVATE KEY-----/g, '').replace(/\n/g, '').replace(/\r/g, '');
        const array = Buffer.from(pem, 'base64');
        const is1024 = array.length === 609 || array.length === 610;
        let pos = is1024 ? 11 : 12;
        const Modulus = this.readBytes(array, pos, is1024 ? 128 : 256); pos += Modulus.length;
        pos += 2;
        const Exponent = this.readBytes(array, pos, 3); pos += 3;
        pos += 4;
        if (array[pos] === 0) pos++;
        const D = this.readBytes(array, pos, is1024 ? 128 : 256);
        return { Modulus, Exponent, D };
    }

    static RsaDecrypt(encryptedData, D, Modulus) {
        const m = BigInteger(encryptedData.toString('hex'), 16);
        const d = BigInteger(D.toString('hex'), 16);
        const n = BigInteger(Modulus.toString('hex'), 16);
        const decrypted = m.modPow(d, n);
        return Buffer.from(decrypted.toString(16).padStart(256, '0'), 'hex');
    }

    static readBytes(buf, pos, len) { return buf.slice(pos, pos + len); }
}

// ====================== DNF 登录注册核心 ======================
class DNFLoginReg {
    constructor() {
        this.ServerIP = '192.168.200.131';
        this.DB_Name = 'game';
        this.DB_Pwd = 'uu5!^%jg';
        this.PRIVATE_KEY = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA2Mzieb9pITcA4f8ZlcJ678aFWWwQURecoJCMUdA5Cg5X6w5/
umgvWmOgadWeZKTehHn08KIObx7L0x/4RerkZFYPdigu7fOoJWUle7i3PU2Ussyx
Qa56O5eAQDuQZGSuKKEqUWL7CFpXJUat0kYg8/6o3U9goWcu4jGokjrUTsrj0++9
3mGssPAeB6NY+LOdkVlw4cQKm7EFmxXeMf3gdnD461PKP6XnlVaLliA+mlEeGpqU
HfQ8S4RlnxGXLP47kgCeU2OA0T1MgRGV8PpduEI5G65TVZ7y7jQrNJZRj3ofVV0t
jB7fNqf/Gj/qyOKBYUWVrZGU2btRHejMVkOnIwIDAQABAoIBAQDTZ6aIFbBMJTiF
J54pPMVoPmsV8ZxPlviyUYGi3aphNe9hVHgUqzdRqshnq1iSx3n8MHg6lawBi0Qy
EYClnREtDgZxr2ljuy8BmAnfBRYZfyc62wMWCy9CIM980xhP7SUulUmQpzYmxQEp
ZixlWOEVTAQaGicd/GHpS4cXYKpaUQ4h0RnjOSut72vR80Dpb2mIM9wkWLPoFO5j
dAsZ4t/CAReO8GgnAPWWFJ1tmZ+wNJCbUN9hp38bUYlDErmkihGzAKGtsSlZHPO7
1cJngaVaJW4igFpiAnkevEfb4kwj/OVQZLdYe1P3wLdkCK+QJu/VCfKC3VzTo+ee
clrRDsnxAoGBAPA4jdxJNoxvOK8mabEiRnh79VEfLc3edOZ+rFK6wN7+S5gj37bd
EnEg2pBpum1XohnObvCQ2UOBuDzKDrWg8WEx++HOdmh8gPXeiImPJhqw3zh2yubc
7cfzNmo+jo2OJAwZXT4LNz/NZY1vBlJjbYdVaWnhaEZYEzU4u2ZMpYO3AoGBAOcK
gCipawANkEO7HfCAZxeTx3f1+JCOH08wJL7agRM8ZA5MP6RwwqpqSNYDlOGjINg/
05AOc6rEM+S+4f91O914qU+32Hf1VqAp65N8zfgaiB6yTl6HjvgZ0qoBtfs+xjoF
rgpSd0ONSzJAsTRfwjG4Xsv3K0VSKwIYjpF4wi/1AoGBAKjlOmiNWTr34ATVny4Z
iS6hCOQWZd/+7nY4zfQEtiKS7Jd1cY9ic9ryXHl4vMiv+prmV33webFK/TxQXHM7
SzspJ42l2f6wuwMjiSAec44EmY1biGE1KEGTMMkWXqgMLjALjVFYFWpYgXQlgW5f
1dx3Ivx8rRH4Ttj2oFvWu/CVAoGBALuDdDyMNAGZAItYJJZ7QDCMigm8on6AOr0E
b+5OXRAFpZdNHyIROo0hMfgwc/cldJTOSKDUeeAQ2aU/nyC8P1gihBflOFUz05iu
LTpIwsoojm2LfbuO/eQy4R2FwfMmIbAZyAUspQs4c91XT/sX9P5xis6zznS2mZ1b
3OoIPmK5AoGAAVUNzBGywcFpR86nmOlXbUJu0aTHEg+J5ktrx+JlVyLOKShVvbKm
e53x8Wu2D/ygtmcGJTXZrQV9VqrJrdUqmGuHVVxg7YjaWwmtrF4Lap5ESBv/WtCK
xjx9KoHY0ZrfXwfrMdPsBAsY6wb2QjpbhI7+8lTV6x+yMEVGTEkpgiU=
-----END RSA PRIVATE KEY-----
        `;
    }

    getMD5(str) {
        return crypto.createHash('md5').update(str, 'ascii').digest('hex');
    }

    async reg(user, pwd) {
        const conn = await mysql.createConnection({
            host: this.ServerIP, user: this.DB_Name, password: this.DB_Pwd, database: 'mysql', ssl: false
        });
        try {
            const [has] = await conn.query('select uid from d_taiwan.accounts where accountname=?', [user]);
            if (has.length) return false;
            let uid = 1;
            const [m] = await conn.query('select uid from d_taiwan.accounts order by uid desc limit 1');
            if (m.length) uid = parseInt(m[0].uid) + 1;
            const md5 = this.getMD5(pwd);
            await conn.query('insert into d_taiwan.accounts (uid,accountname,password) values (?,?,?)', [uid, user, md5]);
            await conn.query('insert into d_taiwan.limit_create_character (m_id) values (?)', [uid]);
            await conn.query('insert into d_taiwan.member_info (m_id,user_id) values (?,?)', [uid, uid]);
            await conn.query('insert into d_taiwan.member_join_info (m_id) values (?)', [uid]);
            await conn.query('insert into d_taiwan.member_miles (m_id) values (?)', [uid]);
            await conn.query('insert into d_taiwan.member_white_account (m_id) values (?)', [uid]);
            await conn.query('insert into taiwan_login.member_login (m_id) values (?)', [uid]);
            await conn.query('insert into taiwan_billing.cash_cera (account,cera,mod_date,reg_date) values (?,100000,now(),now())', [uid]);
            await conn.query('insert into taiwan_billing.cash_cera_point (account,cera_point,reg_date,mod_date) values (?,100000,now(),now())', [uid]);
            await conn.query('insert into taiwan_cain_2nd.member_avatar_coin (m_id) values (?)', [uid]);
            await conn.end();
            return true;
        } catch (e) { await conn.end(); throw e; }
    }

    async login(user, pwd) {
        const conn = await mysql.createConnection({
            host: this.ServerIP, user: this.DB_Name, password: this.DB_Pwd, database: 'mysql', ssl: false
        });
        try {
            const md5 = this.getMD5(pwd);
            const [r] = await conn.query('select uid from d_taiwan.accounts where accountname=? and password=?', [user, md5]);
            if (!r.length) return -1;
            const uid = parseInt(r[0].uid);
            await conn.query('update d_taiwan.limit_create_character set count=0 where m_id=?', [uid]);
            await conn.end();
            return uid;
        } catch (e) { await conn.end(); throw e; }
    }

    getToken(uid) {
        const start = "01FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00";
        const end = "010101010101010101010101010101010101010101010101010101010101010155914510010403030101";
        const hex = start + uid.toString(16).padStart(8, '0') + end;
        const data = DataTranslate.HexStringToByte(hex);
        const key = RSAHelper.ConvertFromPemPrivateKey(this.PRIVATE_KEY);
        const dec = RSAHelper.RsaDecrypt(data, key.D, key.Modulus);
        return dec.toString('base64');
    }
}

const dnf = new DNFLoginReg();

// ====================== 命令行菜单 ======================
function showMenu() {
    console.log("\n===== DOF 登录器 =====");
    console.log("1. 登录游戏");
    console.log("2. 注册账号");
    console.log("======================\n");
    rl.question("请选择操作(1/2)：", async (opt) => {
        if (opt === '1') await doLogin();
        else if (opt === '2') await doReg();
        else {
            console.log("输入错误！");
            showMenu();
        }
    });
}

// 登录
async function doLogin() {
    let user = await question("账号：");
    let pwd = await question("密码：");

    console.log("正在登录...");
    let uid = await dnf.login(user, pwd);

    if (uid < 0) {
        console.log("❌ 账号或密码错误！");
        return showMenu();
    }

    console.log("✅ 登录成功！UID：" + uid);
    let token = dnf.getToken(uid);
    console.log("正在启动游戏...");

    // ✅ 直接运行当前目录的 DNF.exe（不用配置路径）
    exec(`DNF.exe ${token}`, (err) => {
        if (err) console.log("启动失败：", err.message);
        process.exit(0);
        
    });
}

// 注册
async function doReg() {
    let user = await question("账号：");
    let pwd = await question("密码：");

    console.log("正在注册...");
    let ok = await dnf.reg(user, pwd);

    if (ok) {
        console.log("✅ 注册成功！请登录！");
    } else {
        console.log("❌ 账号已存在！");
    }
    showMenu();
}

// 输入工具
function question(msg) {
    return new Promise(res => rl.question(msg, res));
}

// 启动
showMenu();