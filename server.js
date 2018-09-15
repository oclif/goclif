const net = require('net')
const fs = require('fs')
const heroku = require('heroku')

try {
  // remove existing socket if exists
  fs.unlinkSync('/tmp/foo.sock')
} catch {}

const stdout = msg => stdout.write(msg)
stdout.write = process.stdout.write.bind(process.stdout)

const stderr = msg => stderr.write(msg)
stderr.write = process.stderr.write.bind(process.stderr)

net
  .createServer()
  .listen('/tmp/foo.sock', () => {
    console.log('server listening')
  })
  .on('connection', socket => {
    console.log('server listening')
    socket.setEncoding('utf8')
    socket.on('data', data => {
      console.log(`server received: ${data}`)
      const send = msg => {
        stdout(`server sent: ${msg}`)
        socket.write(msg)
      }
      process.stdout.write = d => {
        send(d)
      }
      process.stderr.write = d => {
        send(d)
      }
      heroku.run([data])
      .then(() => {
        socket.write(data.toUpperCase())
        socket.end()
      })
      .catch(err => {
        if (err.code === 'EEXIT') {
          send(`EEXIT: ${err.oclif.exit}`)
          socket.end()
        } else {
          console.error(err)
        }
      })
    })
  })
