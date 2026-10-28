import 'package:quickeats_customer/core/constants/api_constants.dart';
import 'package:quickeats_customer/core/network/api_client.dart';
import 'package:quickeats_customer/models/payment_card.dart';

class CardService {
  final ApiClient _api;
  CardService({required ApiClient apiClient}) : _api = apiClient;

  /// The saved card, or null if the customer hasn't added one.
  Future<PaymentCard?> getCard() async {
    final resp = await _api.dio.get(ApiConstants.myCard);
    final data = resp.data['data'];
    return data is Map<String, dynamic> ? PaymentCard.fromJson(data) : null;
  }

  /// Adds a card, replacing the current one. Throws a DioException with a field-level
  /// message (error.details.field) when the server rejects the details.
  Future<PaymentCard> addCard({required String number, required String holderName, required int expMonth, required int expYear, required String cvv}) async {
    final resp = await _api.dio.put(ApiConstants.myCard, data: {
      'number': number, 'holderName': holderName, 'expMonth': expMonth, 'expYear': expYear, 'cvv': cvv,
    });
    return PaymentCard.fromJson(resp.data['data']);
  }

  Future<void> removeCard() async {
    await _api.dio.delete(ApiConstants.myCard);
  }

  Future<PayCheck> canPay(int amountCents) async {
    final resp = await _api.dio.get(ApiConstants.myCardCheck, queryParameters: {'amount': amountCents});
    return PayCheck.fromJson(resp.data['data']);
  }
}
