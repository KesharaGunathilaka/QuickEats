import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/core/constants/api_constants.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

typedef WSMessageHandler = void Function(Map<String, dynamic> message);

/// Returns a valid access token, or null when the user is logged out.
typedef TokenProvider = Future<String?> Function();

/// Connects to the notification service and authenticates as the customer.
/// Reconnects automatically, asking for a fresh token each time.
class WSClient {
  WebSocketChannel? _channel;
  StreamSubscription? _subscription;
  final List<WSMessageHandler> _handlers = [];
  TokenProvider? _tokenProvider;
  bool _shouldReconnect = false;
  Timer? _reconnectTimer;

  bool get isConnected => _channel != null;

  void connect(TokenProvider tokenProvider) {
    _tokenProvider = tokenProvider;
    _shouldReconnect = true;
    _doConnect();
  }

  Future<void> _doConnect() async {
    final token = await _tokenProvider?.call();
    if (token == null || !_shouldReconnect) return;

    try {
      _channel = WebSocketChannel.connect(Uri.parse(ApiConstants.wsUrl));
      _channel!.sink.add(jsonEncode({'type': 'AUTH', 'token': token}));

      _subscription = _channel!.stream.listen(
        (data) {
          try {
            final message = jsonDecode(data as String) as Map<String, dynamic>;
            for (final handler in List.of(_handlers)) {
              handler(message);
            }
          } catch (e) {
            debugPrint('[WS] Parse error: $e');
          }
        },
        onError: (error) {
          debugPrint('[WS] Error: $error');
          _scheduleReconnect();
        },
        onDone: () {
          debugPrint('[WS] Connection closed');
          _scheduleReconnect();
        },
      );
      debugPrint('[WS] Connecting to ${ApiConstants.wsUrl}');
    } catch (e) {
      debugPrint('[WS] Connect failed: $e');
      _scheduleReconnect();
    }
  }

  void _scheduleReconnect() {
    _subscription?.cancel();
    _channel = null;
    if (!_shouldReconnect) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(const Duration(seconds: 3), _doConnect);
  }

  void addHandler(WSMessageHandler handler) => _handlers.add(handler);

  void removeHandler(WSMessageHandler handler) => _handlers.remove(handler);

  void disconnect() {
    _shouldReconnect = false;
    _reconnectTimer?.cancel();
    _subscription?.cancel();
    _channel?.sink.close();
    _channel = null;
  }
}
