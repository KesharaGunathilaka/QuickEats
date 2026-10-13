import 'package:flutter/foundation.dart';
import 'package:quickeats_shop/models/menu_item.dart';
import 'package:quickeats_shop/services/auth_service.dart';
import 'package:quickeats_shop/services/menu_service.dart';

class MenuViewModel extends ChangeNotifier {
  final MenuService _menuService;
  final AuthService _authService;

  MenuViewModel({required MenuService menuService, required AuthService authService})
    : _menuService = menuService, _authService = authService;

  List<MenuItem> _items = [];
  bool _isLoading = false;
  String? _error;

  List<MenuItem> get items => _items;
  bool get isLoading => _isLoading;
  String? get error => _error;
  String? get shopId => _authService.shopId;

  Map<String, List<MenuItem>> get itemsByCategory {
    final map = <String, List<MenuItem>>{};
    for (final item in _items) {
      map.putIfAbsent(item.category, () => []).add(item);
    }
    return map;
  }

  Future<void> loadItems() async {
    final sid = shopId;
    if (sid == null) return;

    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _items = await _menuService.getMenuItems(sid);
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _isLoading = false;
      _error = 'Failed to load menu: $e';
      notifyListeners();
    }
  }

  Future<bool> addItem({required String name, required String description, required int price, required String category, required int stock}) async {
    final sid = shopId;
    if (sid == null) return false;

    try {
      final item = await _menuService.createItem(sid, name: name, description: description, price: price, category: category, stock: stock);
      _items.add(item);
      notifyListeners();
      return true;
    } catch (e) {
      _error = 'Failed to add item: $e';
      notifyListeners();
      return false;
    }
  }

  Future<bool> updateStock(String itemId, int stock) async {
    try {
      await _menuService.updateStock(itemId, stock);
      final idx = _items.indexWhere((i) => i.id == itemId);
      if (idx >= 0) {
        _items[idx].stock = stock;
        _items[idx].available = stock > 0;
        notifyListeners();
      }
      return true;
    } catch (e) {
      _error = 'Failed to update stock: $e';
      notifyListeners();
      return false;
    }
  }

  Future<bool> deleteItem(String itemId) async {
    try {
      await _menuService.deleteItem(itemId);
      _items.removeWhere((i) => i.id == itemId);
      notifyListeners();
      return true;
    } catch (e) {
      _error = 'Failed to delete item: $e';
      notifyListeners();
      return false;
    }
  }
}
