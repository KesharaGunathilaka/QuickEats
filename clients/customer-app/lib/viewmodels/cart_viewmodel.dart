import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/models/cart_item.dart';
import 'package:quickeats_customer/models/menu_item.dart';

class CartViewModel extends ChangeNotifier {
  final List<CartItem> _items = [];
  String? _shopId;
  String? _shopName;

  List<CartItem> get items => List.unmodifiable(_items);
  String? get shopId => _shopId;
  String? get shopName => _shopName;
  int get itemCount => _items.fold(0, (sum, item) => sum + item.quantity);
  int get totalCents => _items.fold(0, (sum, item) => sum + item.subtotal);
  String get formattedTotal => 'LKR ${(totalCents / 100).toStringAsFixed(2)}';
  bool get isEmpty => _items.isEmpty;

  void addItem(MenuItem menuItem, {String? shopId, String? shopName}) {
    // If adding from a different shop, clear cart
    if (_shopId != null && _shopId != (shopId ?? menuItem.shopId)) {
      _items.clear();
    }
    _shopId = shopId ?? menuItem.shopId;
    _shopName = shopName;

    final idx = _items.indexWhere((c) => c.menuItem.id == menuItem.id);
    if (idx >= 0) { _items[idx].quantity++; } else { _items.add(CartItem(menuItem: menuItem)); }
    notifyListeners();
  }

  void removeItem(String menuItemId) {
    _items.removeWhere((c) => c.menuItem.id == menuItemId);
    if (_items.isEmpty) { _shopId = null; _shopName = null; }
    notifyListeners();
  }

  void updateQuantity(String menuItemId, int qty) {
    if (qty <= 0) { removeItem(menuItemId); return; }
    final idx = _items.indexWhere((c) => c.menuItem.id == menuItemId);
    if (idx >= 0) { _items[idx].quantity = qty; notifyListeners(); }
  }

  void clear() { _items.clear(); _shopId = null; _shopName = null; notifyListeners(); }
}
