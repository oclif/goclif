const net = require('net')
const fs = require('fs')
const heroku = require('heroku')
console.dir(heroku)

try {
  // remove existing socket if exists
  fs.unlinkSync('/tmp/foo.sock')
} catch {}

net
  .createServer()
  .listen('/tmp/foo.sock', () => {
    console.log('server listening')
  })
  .on('connection', socket => {
    console.log('server listening')
    socket.setEncoding('utf8')
    socket.on('data', data => {
      heroku.run(['version'])
      console.log(`server received: ${data}`)
      socket.write(data.toUpperCase())
      socket.end()
    })
  })
