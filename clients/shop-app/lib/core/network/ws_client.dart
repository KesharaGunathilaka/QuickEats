import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:quickeats_shop/core/constants/api_constants.dart';

typedef WSMessageHandler = void Function(Map<String, dynamic> message);

class WSClient {
  WebSocketChannel? _channel;
  StreamSubscription? _subscription;
  final List<WSMessageHandler> _handlers = [];
  String? _token;
  String? _shopId;
  bool _shouldReconnect = true;
  Timer? _reconnectTimer;

  void connect({required String token, required String shopId}) {
    _token = token;
    _shopId = shopId;
    _shouldReconnect = true;
    _doConnect();
  }

  void _doConnect() {
    try {
      _channel = WebSocketChannel.connect(
        Uri.parse('${ApiConstants.wsUrl}?shopId=$_shopId'),
      );

      // Send AUTH message
      _channel!.sink.add(jsonEncode({
        'type': 'AUTH',
        'token': _token,
      }));

      _subscription = _channel!.stream.listen(
        (data) {
          try {
            final message = jsonDecode(data as String) as Map<String, dynamic>;
            for (final handler in _handlers) {
              handler(message);
            }
          } catch (e) {
            print('[WS] Parse error: $e');
          }
        },
        onError: (error) {
          print('[WS] Error: $error');
          _scheduleReconnect();
        },
        onDone: () {
          print('[WS] Connection closed');
          _scheduleReconnect();
        },
      );

      print('[WS] Connected to ${ApiConstants.wsUrl}');
    } catch (e) {
      print('[WS] Connect failed: $e');
      _scheduleReconnect();
    }
  }

  void _scheduleReconnect() {
    if (!_shouldReconnect || _token == null) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(const Duration(seconds: 3), () {
      print('[WS] Reconnecting...');
      _doConnect();
    });
  }

  void addHandler(WSMessageHandler handler) {
    _handlers.add(handler);
  }

  void removeHandler(WSMessageHandler handler) {
    _handlers.remove(handler);
  }

  void disconnect() {
    _shouldReconnect = false;
    _reconnectTimer?.cancel();
    _subscription?.cancel();
    _channel?.sink.close();
    _channel = null;
  }
}
