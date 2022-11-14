package errors

import "errors"

var ErrRentDatesUsed = errors.New("as datas já estao em uso")
var ErrRentReversedDates = errors.New("a data de início deve anteceder a data de fim da estadia")
var ErrCondoAlreadyPayed = errors.New("o condomínio do mês informado já foi pago")
var ErrBillAlreadyPayed = errors.New("a conta de energia já foi paga nesse mês")
var ErrCleaningAlreadyHappened = errors.New("uma faxina já foi cadastrada nesse mesmo dia")
var ErrMiscellaneousExpenseAlreadyCreated = errors.New("essa despesa já foi adicionada previamente")
