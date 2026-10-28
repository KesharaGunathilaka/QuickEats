import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/models/payment_card.dart';
import 'package:quickeats_customer/services/account_service.dart';
import 'package:quickeats_customer/services/card_service.dart';

/// What checkout can tell the customer about paying for the current cart.
enum PaymentReadiness { loading, noCard, insufficient, ready, unknown }

class CardViewModel extends ChangeNotifier {
  final CardService _cardService;
  CardViewModel({required CardService cardService}) : _cardService = cardService;

  PaymentCard? _card;
  bool _isLoading = false;
  bool _loaded = false;
  bool _isSaving = false;
  String? _error;
  Map<String, String> _fieldErrors = {};
  PaymentReadiness _readiness = PaymentReadiness.loading;
  int? _checkedAmount;

  PaymentCard? get card => _card;
  bool get hasCard => _card != null;
  bool get isLoading => _isLoading;
  bool get loaded => _loaded;
  bool get isSaving => _isSaving;
  String? get error => _error;
  /// Server-side validation message per field: number, holderName, expiry, cvv.
  Map<String, String> get fieldErrors => _fieldErrors;
  PaymentReadiness get readiness => _readiness;

  Future<void> loadCard() async {
    _isLoading = true; _error = null; notifyListeners();
    try {
      _card = await _cardService.getCard();
      _loaded = true;
    } catch (e) {
      _error = apiErrorMessage(e, 'Couldn\'t load your card');
    }
    _isLoading = false; notifyListeners();
  }

  /// Works out whether the saved card can pay amountCents (the balance itself is
  /// never sent to the app).
  Future<void> checkPayment(int amountCents) async {
    _checkedAmount = amountCents;
    _readiness = PaymentReadiness.loading; notifyListeners();
    try {
      _card = await _cardService.getCard();
      _loaded = true;
      if (_card == null) {
        _readiness = PaymentReadiness.noCard;
      } else {
        final check = await _cardService.canPay(amountCents);
        _readiness = !check.hasCard
            ? PaymentReadiness.noCard
            : check.sufficient ? PaymentReadiness.ready : PaymentReadiness.insufficient;
      }
    } catch (_) {
      _readiness = PaymentReadiness.unknown;
    }
    if (_checkedAmount == amountCents) notifyListeners();
  }

  /// Marks the result of an order the server refused (codes NO_CARD / INSUFFICIENT_FUNDS).
  void markDeclined(String code) {
    if (code == 'NO_CARD') { _card = null; _readiness = PaymentReadiness.noCard; }
    if (code == 'INSUFFICIENT_FUNDS') _readiness = PaymentReadiness.insufficient;
    notifyListeners();
  }

  Future<bool> addCard({required String number, required String holderName, required int expMonth, required int expYear, required String cvv}) async {
    _isSaving = true; _error = null; _fieldErrors = {}; notifyListeners();
    try {
      _card = await _cardService.addCard(number: number, holderName: holderName, expMonth: expMonth, expYear: expYear, cvv: cvv);
      _loaded = true;
      _isSaving = false; notifyListeners();
      return true;
    } catch (e) {
      final field = _errorField(e);
      final msg = apiErrorMessage(e, 'Couldn\'t add the card');
      if (field != null) { _fieldErrors = {field: msg}; } else { _error = msg; }
      _isSaving = false; notifyListeners();
      return false;
    }
  }

  Future<bool> removeCard() async {
    try {
      await _cardService.removeCard();
      _card = null;
      _readiness = PaymentReadiness.noCard;
      notifyListeners();
      return true;
    } catch (e) {
      _error = apiErrorMessage(e, 'Couldn\'t remove the card');
      notifyListeners();
      return false;
    }
  }

  void clearErrors() { _error = null; _fieldErrors = {}; notifyListeners(); }

  String? _errorField(Object e) {
    if (e is DioException) {
      final data = e.response?.data;
      if (data is Map && data['error'] is Map && data['error']['details'] is Map) {
        final field = data['error']['details']['field'];
        if (field is String) return field;
      }
    }
    return null;
  }
}
