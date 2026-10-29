// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IProviderPayoutStorage {
    /**
     * The event that is emitted when the provider payout target updated.
     * @param provider The provider address.
     * @param payoutTarget The effective payout target, the provider address when reset.
     */
    event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget);

    /**
     * The event that is emitted when the provider reward accrued.
     * @param sessionId The session ID.
     * @param provider The provider address.
     * @param accrued The accrued amount, equal to `paid + owed`.
     * @param paid The amount paid to the payout target now.
     * @param owed The amount recorded as provider debt.
     */
    event ProviderRewardAccrued(
        bytes32 indexed sessionId,
        address indexed provider,
        uint256 accrued,
        uint256 paid,
        uint256 owed
    );

    /**
     * The event that is emitted when the provider debt paid.
     * @param provider The provider address.
     * @param payoutTarget The payout target that received the payment.
     * @param amount The paid amount.
     */
    event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount);

}
