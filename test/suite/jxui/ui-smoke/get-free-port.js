var net = require('net'),
    srv = net.createServer(() => {});

srv.listen(0, () => {
    console.log(srv.address().port);
    process.exit();
});
