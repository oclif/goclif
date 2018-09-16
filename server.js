const net = require('net')
const fs = require('fs')
const heroku = require('heroku')
const {inspect} = require('util')

try {
  // remove existing socket if exists
  fs.unlinkSync('/tmp/foo.sock')
} catch {}

const stdout = msg => stdout.write(msg)
stdout.write = process.stdout.write.bind(process.stdout)

const stderr = msg => stderr.write(msg)
stderr.write = process.stderr.write.bind(process.stderr)

const debug = msg => stderr(msg + '\n')

const server = net.createServer()
const socket = '/tmp/foo.sock'
server.listen(socket, () => {
  debug('server listening')
  stdout(socket + '\n')
})
server.on('connection', socket => {
  debug('socket connected')
  socket.setEncoding('utf8')
  socket.on('data', data => {
    data = JSON.parse(data)
    debug(`server received: ${inspect(data)}`)
    const send = msg => {
      debug(`server sent: ${inspect(msg)}`)
      socket.write(msg)
    }
    const end = () => {
      debug('server closing socket\n')
      socket.end()
    }
    process.stdout.write = d => {
      send(d)
    }
    process.stderr.write = d => {
      send(d)
    }
    heroku.run(data.argv)
    .then(() => {
      socket.write(data.toUpperCase())
      socket.end()
    })
    .catch(err => {
      if (err.code === 'EEXIT') {
        send(`EEXIT: ${err.oclif.exit}`)
        end()
      } else {
        stderr(err.message)
      }
    })
  })
})

server.on('close', () => {
  debug('server closed')
  process.exit(0)
})

setTimeout(() => {
  debug('server timed out')
  server.close()
}, 10000)
