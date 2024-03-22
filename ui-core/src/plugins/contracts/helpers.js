//@ts-check

const remove0xPrefix = privateKey => privateKey.replace('0x', '');

// https://superuser.com/a/1465498
const add65BytesPrefix = key => `04${key}`;

module.exports = {
    remove0xPrefix,
    add65BytesPrefix,
}