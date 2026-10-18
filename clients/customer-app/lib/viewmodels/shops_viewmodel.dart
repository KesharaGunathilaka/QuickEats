import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/models/shop.dart';
import 'package:quickeats_customer/models/menu_item.dart';
import 'package:quickeats_customer/services/shop_service.dart';

class ShopsViewModel extends ChangeNotifier {
  final ShopService _shopService;
  ShopsViewModel({required ShopService shopService}) : _shopService = shopService;

  List<Shop> _shops = [];
  List<MenuItem> _menuItems = [];
  Shop? _selectedShop;
  bool _isLoading = false;
  String? _error;

  List<Shop> get shops => _shops;
  List<MenuItem> get menuItems => _menuItems;
  Shop? get selectedShop => _selectedShop;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Map<String, List<MenuItem>> get menuByCategory {
    final m = <String, List<MenuItem>>{};
    for (final item in _menuItems) m.putIfAbsent(item.category, () => []).add(item);
    return m;
  }

  Future<void> loadShops() async {
    _isLoading = true; _error = null; notifyListeners();
    try { _shops = await _shopService.getShops(); } catch (e) { _error = 'Failed to load shops'; }
    _isLoading = false; notifyListeners();
  }

  Future<void> selectShop(String shopId) async {
    _isLoading = true; notifyListeners();
    try {
      _selectedShop = await _shopService.getShop(shopId);
      _menuItems = await _shopService.getMenu(shopId);
    } catch (e) { _error = 'Failed to load menu'; }
    _isLoading = false; notifyListeners();
  }
}
