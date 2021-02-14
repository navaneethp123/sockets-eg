import 'dart:convert';
import 'dart:math';

import 'package:http/http.dart' as http;
import 'package:logging/logging.dart';
import 'package:web_socket_channel/io.dart';

final baseUrl = 'http://localhost:5000';

void main() async {
  final log = Logger('main');
  log.onRecord.listen((record) {
    print('${record.time} ${record.message}');
  });

  final sckt = IOWebSocketChannel.connect('ws://localhost:5000/pushnotifs');
  sckt.stream.listen((data) {
    log.info('server: $data');
  }, onError: (err) {
    log.warning('error: $err');
  }, onDone: () {
    log.info('server closed.');
  });

  final menuRes = await http.get(Uri.parse('$baseUrl/menu'));
  final menu = (jsonDecode(menuRes.body) as List).cast<Map<String, dynamic>>();
  
  final rand = Random();
  
  while (true) {
    await Future.delayed(Duration(seconds: 10));

    final table = rand.nextInt(10) + 1;
    final numOrders = rand.nextInt(7) + 1;

    final items = <int>{};

    var i = 0;
    while (i < numOrders) {
      final itemId = menu[rand.nextInt(menu.length)]['id'];
      if (items.contains(itemId)) {
        continue;
      }

      items.add(itemId);
      ++i;
    }

    final order = jsonEncode({
      'tableNo': table,
      'items': items.map((id) => {
        'itemID': id,
        'quantity': rand.nextInt(5) + 1,
      }).toList(),
    });

    log.info('created order: $order');

    final orderRes = await http.post('$baseUrl/order', body: order);
    log.info('orderRes status code: ${orderRes.statusCode}');
  }
}
